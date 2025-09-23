package wavespeed

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"one-api/dto"
	"one-api/types"
	"time"
)

// WavespeedResponse represents the standard WaveSpeed API response
type WavespeedResponse struct {
	ID               string                 `json:"id"`
	Status           string                 `json:"status"`
	CreatedAt        string                 `json:"created_at"`
	Model            string                 `json:"model"`
	Outputs          []string               `json:"outputs"`
	HasNSFWContents  []bool                 `json:"has_nsfw_contents"`
	Error            string                 `json:"error,omitempty"`
	ErrorCode        string                 `json:"error_code,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	AdditionalFields map[string]interface{} `json:"-"` // For any additional fields
}

// WavespeedErrorResponse represents error response from WaveSpeed API
type WavespeedErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// ResponseHandler handles WaveSpeed API responses
type ResponseHandler struct{}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// HandleResponse processes WaveSpeed API response and converts to OpenAI format
func (rh *ResponseHandler) HandleResponse(resp *http.Response, modelPath string) (*dto.OpenAITextResponse, *types.NewAPIError) {
	if resp.StatusCode != http.StatusOK {
		return nil, rh.handleErrorResponse(resp)
	}

	// Parse WaveSpeed response
	wavespeedResp, err := rh.parseWavespeedResponse(resp)
	if err != nil {
		return nil, types.NewI18nError(context.Background(), err, types.ErrorCodeDoRequestFailed)
	}

	// Check for API-level errors
	if wavespeedResp.Error != "" {
		return nil, types.NewI18nError(context.Background(), errors.New(wavespeedResp.Error), types.ErrorCodeDoRequestFailed)
	}

	// Convert to OpenAI format based on model type
	if IsVideoModel(modelPath) {
		return rh.convertVideoResponse(wavespeedResp)
	} else {
		return rh.convertImageResponse(wavespeedResp)
	}
}

// parseWavespeedResponse parses the WaveSpeed API response
func (rh *ResponseHandler) parseWavespeedResponse(resp *http.Response) (*WavespeedResponse, error) {
	var wavespeedResp WavespeedResponse
	err := json.NewDecoder(resp.Body).Decode(&wavespeedResp)
	if err != nil {
		return nil, err
	}
	return &wavespeedResp, nil
}

// convertImageResponse converts WaveSpeed image response to OpenAI format
func (rh *ResponseHandler) convertImageResponse(wavespeedResp *WavespeedResponse) (*dto.OpenAITextResponse, *types.NewAPIError) {
	// For image generation, we typically return the image URLs
	// Convert to a text response that describes the generated images
	var content string
	if len(wavespeedResp.Outputs) > 0 {
		content = "Image generated successfully. URLs: "
		for i, url := range wavespeedResp.Outputs {
			if i > 0 {
				content += ", "
			}
			content += url
		}
	} else {
		content = "Image generation completed but no outputs available"
	}

	// Check task status
	if wavespeedResp.Status == "processing" || wavespeedResp.Status == "created" {
		content = "Image generation is in progress. Task ID: " + wavespeedResp.ID
	} else if wavespeedResp.Status == "failed" {
		return nil, types.NewI18nError(context.Background(), errors.New("image generation failed"), types.ErrorCodeDoRequestFailed)
	}

	openaiResponse := &dto.OpenAITextResponse{
		Object: "chat.completion",
		Id:     wavespeedResp.ID,
		Model:  wavespeedResp.Model,
		Choices: []dto.OpenAITextResponseChoice{
			{
				Index: 0,
				Message: dto.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: rh.mapStatusToFinishReason(wavespeedResp.Status),
			},
		},
		Usage: dto.Usage{
			PromptTokens:     rh.estimatePromptTokens(wavespeedResp),
			CompletionTokens: rh.estimateCompletionTokens(wavespeedResp),
			TotalTokens:      0, // Will be calculated
		},
		Created: rh.parseCreatedAt(wavespeedResp.CreatedAt),
	}

	// Calculate total tokens
	openaiResponse.Usage.TotalTokens = openaiResponse.Usage.PromptTokens + openaiResponse.Usage.CompletionTokens

	return openaiResponse, nil
}

// convertVideoResponse converts WaveSpeed video response to OpenAI format
func (rh *ResponseHandler) convertVideoResponse(wavespeedResp *WavespeedResponse) (*dto.OpenAITextResponse, *types.NewAPIError) {
	// For video generation, we return the video URLs
	var content string
	if len(wavespeedResp.Outputs) > 0 {
		content = "Video generated successfully. URLs: "
		for i, url := range wavespeedResp.Outputs {
			if i > 0 {
				content += ", "
			}
			content += url
		}
	} else {
		content = "Video generation completed but no outputs available"
	}

	// Check task status
	if wavespeedResp.Status == "processing" || wavespeedResp.Status == "created" {
		content = "Video generation is in progress. Task ID: " + wavespeedResp.ID
	} else if wavespeedResp.Status == "failed" {
		return nil, types.NewI18nError(context.Background(), errors.New("video generation failed"), types.ErrorCodeDoRequestFailed)
	}

	openaiResponse := &dto.OpenAITextResponse{
		Object: "chat.completion",
		Id:     wavespeedResp.ID,
		Model:  wavespeedResp.Model,
		Choices: []dto.OpenAITextResponseChoice{
			{
				Index: 0,
				Message: dto.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: rh.mapStatusToFinishReason(wavespeedResp.Status),
			},
		},
		Usage: dto.Usage{
			PromptTokens:     rh.estimatePromptTokens(wavespeedResp),
			CompletionTokens: rh.estimateCompletionTokens(wavespeedResp),
			TotalTokens:      0, // Will be calculated
		},
		Created: rh.parseCreatedAt(wavespeedResp.CreatedAt),
	}

	// Calculate total tokens
	openaiResponse.Usage.TotalTokens = openaiResponse.Usage.PromptTokens + openaiResponse.Usage.CompletionTokens

	return openaiResponse, nil
}

// handleErrorResponse handles error responses from WaveSpeed API
func (rh *ResponseHandler) handleErrorResponse(resp *http.Response) *types.NewAPIError {
	var errorResp WavespeedErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errorResp)
	if err != nil {
		return types.NewI18nError(context.Background(), errors.New("failed to parse error response"), types.ErrorCodeReadResponseBodyFailed)
	}

	return types.NewI18nError(context.Background(), errors.New(errorResp.Error.Message), rh.mapHTTPStatusToErrorCode(resp.StatusCode))
}

// Helper methods

func (rh *ResponseHandler) mapStatusToFinishReason(status string) string {
	switch status {
	case "completed":
		return "stop"
	case "failed":
		return "stop"
	case "processing", "created":
		return "length" // Still processing
	default:
		return "stop"
	}
}

func (rh *ResponseHandler) mapHTTPStatusToErrorCode(statusCode int) types.ErrorCode {
	switch statusCode {
	case http.StatusBadRequest:
		return types.ErrorCodeInvalidRequest
	case http.StatusUnauthorized:
		return types.ErrorCodeChannelInvalidKey
	case http.StatusForbidden:
		return types.ErrorCodeAccessDenied
	case http.StatusTooManyRequests:
		return types.ErrorCodeChannelResponseTimeExceeded
	case http.StatusInternalServerError:
		return types.ErrorCodeDoRequestFailed
	default:
		return types.ErrorCodeDoRequestFailed
	}
}

func (rh *ResponseHandler) estimatePromptTokens(resp *WavespeedResponse) int {
	// For image/video generation, we estimate based on the complexity
	// This is a rough estimation since WaveSpeed doesn't provide token counts
	baseTokens := 50 // Base tokens for the request overhead

	// Add estimated tokens based on outputs (more complex generations likely had more detailed prompts)
	if len(resp.Outputs) > 0 {
		baseTokens += len(resp.Outputs) * 20
	}

	return baseTokens
}

func (rh *ResponseHandler) estimateCompletionTokens(resp *WavespeedResponse) int {
	// For image/video generation, completion tokens represent the generated content
	// We estimate based on the number of outputs
	if len(resp.Outputs) > 0 {
		return len(resp.Outputs) * 100 // Rough estimate per generated item
	}
	return 50 // Default for the response text
}

func (rh *ResponseHandler) parseCreatedAt(createdAt string) int64 {
	// Parse ISO timestamp to Unix timestamp
	if createdAt == "" {
		return time.Now().Unix()
	}

	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return time.Now().Unix()
	}

	return t.Unix()
}

// IsTaskComplete checks if a task is complete based on status
func (rh *ResponseHandler) IsTaskComplete(status string) bool {
	return status == "completed" || status == "failed"
}

// GetTaskResult extracts the final result from a completed task
func (rh *ResponseHandler) GetTaskResult(wavespeedResp *WavespeedResponse) ([]string, error) {
	if wavespeedResp.Status == "failed" {
		return nil, errors.New("task failed: " + wavespeedResp.Error)
	}

	if wavespeedResp.Status != "completed" {
		return nil, errors.New("task not completed yet")
	}

	return wavespeedResp.Outputs, nil
}