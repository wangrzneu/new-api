package wavespeed

import (
	"encoding/json"
	"errors"
	"one-api/dto"
	"strings"
)

// RequestConverter handles converting different request types to WaveSpeed format
type RequestConverter struct {
	schemaManager *SchemaManager
}

// NewRequestConverter creates a new request converter
func NewRequestConverter() *RequestConverter {
	return &RequestConverter{
		schemaManager: GetSchemaManager(),
	}
}

// ConvertImageRequest converts OpenAI image request to WaveSpeed format using schema
func (rc *RequestConverter) ConvertImageRequest(request dto.ImageRequest, modelPath string) (map[string]interface{}, error) {
	if request.Prompt == "" {
		return nil, errors.New("prompt is required")
	}

	// Extract parameters from the request
	params := make(map[string]interface{})
	params["prompt"] = request.Prompt

	// Handle size parameter (convert to aspect_ratio or size based on model)
	if request.Size != "" {
		if isAspectRatioModel(modelPath) {
			params["aspect_ratio"] = convertSizeToAspectRatio(request.Size)
		} else {
			params["size"] = request.Size
		}
	}

	// Handle number of images
	if request.N > 0 {
		params["num_images"] = request.N
	}

	// Handle quality parameter (map to model-specific parameters)
	if request.Quality != "" {
		params = rc.handleQualityParameter(params, request.Quality, modelPath)
	}

	// Handle style parameter
	if len(request.Style) > 0 {
		var style string
		if err := json.Unmarshal(request.Style, &style); err == nil {
			params = rc.handleStyleParameter(params, style, modelPath)
		}
	}

	// Build request using schema
	return rc.schemaManager.BuildRequestFromSchema(modelPath, params)
}

// ConvertVideoRequest converts image request to video format (for I2V models)
func (rc *RequestConverter) ConvertVideoRequest(request dto.ImageRequest, modelPath string) (map[string]interface{}, error) {
	// Extract parameters from the request
	params := make(map[string]interface{})

	if request.Prompt != "" {
		params["prompt"] = request.Prompt
	}

	// For video models, try to extract image from the request
	// This could come from multipart upload or base64 encoding
	imageURL := rc.extractImageFromRequest(request)
	if imageURL != "" {
		params["image"] = imageURL
	}

	// Handle size for video models
	if request.Size != "" {
		params["size"] = request.Size
	}

	// Build request using schema
	return rc.schemaManager.BuildRequestFromSchema(modelPath, params)
}

// ConvertOpenAIRequest converts general OpenAI request to WaveSpeed format
func (rc *RequestConverter) ConvertOpenAIRequest(request *dto.GeneralOpenAIRequest, modelPath string) (map[string]interface{}, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// Extract parameters from messages
	params := make(map[string]interface{})

	// Extract prompt from messages
	prompt := rc.extractPromptFromMessages(request.Messages)
	if prompt != "" {
		params["prompt"] = prompt
	}

	// Extract image from messages (for I2V models)
	imageURL := rc.extractImageFromMessages(request.Messages)
	if imageURL != "" {
		params["image"] = imageURL
	}

	// Handle temperature (map to guidance_scale if applicable)
	if request.Temperature != nil {
		params = rc.handleTemperatureParameter(params, *request.Temperature, modelPath)
	}

	// Handle max_tokens (not directly applicable to image/video generation)

	// Build request using schema
	return rc.schemaManager.BuildRequestFromSchema(modelPath, params)
}

// Helper methods

func (rc *RequestConverter) extractPromptFromMessages(messages []dto.Message) string {
	var prompts []string
	for _, message := range messages {
		if message.Content != nil {
			switch v := message.Content.(type) {
			case string:
				prompts = append(prompts, v)
			case []interface{}:
				for _, item := range v {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if text, exists := itemMap["text"]; exists {
							if textStr, ok := text.(string); ok {
								prompts = append(prompts, textStr)
							}
						}
					}
				}
			}
		}
	}
	return strings.Join(prompts, " ")
}

func (rc *RequestConverter) extractImageFromMessages(messages []dto.Message) string {
	for _, message := range messages {
		if message.Content != nil {
			if contentSlice, ok := message.Content.([]interface{}); ok {
				for _, item := range contentSlice {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if itemType, exists := itemMap["type"]; exists && itemType == "image_url" {
							if imageUrl, exists := itemMap["image_url"]; exists {
								if urlMap, ok := imageUrl.(map[string]interface{}); ok {
									if url, exists := urlMap["url"]; exists {
										if urlStr, ok := url.(string); ok {
											return urlStr
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func (rc *RequestConverter) extractImageFromRequest(request dto.ImageRequest) string {
	// This would be implemented based on how images are passed in the request
	// Could be from form data, base64 encoding, etc.
	return ""
}

func (rc *RequestConverter) handleQualityParameter(params map[string]interface{}, quality, modelPath string) map[string]interface{} {
	// Map quality to model-specific parameters
	switch quality {
	case "hd":
		if strings.Contains(modelPath, "flux") {
			params["num_inference_steps"] = 50
		}
	case "standard":
		if strings.Contains(modelPath, "flux") {
			params["num_inference_steps"] = 28
		}
	}
	return params
}

func (rc *RequestConverter) handleStyleParameter(params map[string]interface{}, style, modelPath string) map[string]interface{} {
	// Map style to model-specific parameters or modify prompt
	switch style {
	case "vivid":
		if prompt, exists := params["prompt"]; exists {
			params["prompt"] = prompt.(string) + ", vivid colors, high contrast"
		}
	case "natural":
		if prompt, exists := params["prompt"]; exists {
			params["prompt"] = prompt.(string) + ", natural lighting, realistic"
		}
	}
	return params
}

func (rc *RequestConverter) handleTemperatureParameter(params map[string]interface{}, temperature float64, modelPath string) map[string]interface{} {
	// Map temperature to guidance_scale for models that support it
	if strings.Contains(modelPath, "flux") || strings.Contains(modelPath, "seedream") {
		// Convert temperature (0-2) to guidance_scale (typically 1-20)
		guidanceScale := 1.0 + (temperature * 9.5) // Maps 0-2 to 1-20
		params["guidance_scale"] = guidanceScale
	}
	return params
}

func isAspectRatioModel(modelPath string) bool {
	aspectRatioModels := []string{
		"wavespeed-ai/flux-1.1-pro",
		"google/imagen3",
		"google/imagen4",
	}

	for _, model := range aspectRatioModels {
		if strings.Contains(modelPath, model) {
			return true
		}
	}
	return false
}

func convertSizeToAspectRatio(size string) string {
	sizeMap := map[string]string{
		"1024x1024": "1:1",
		"1024x768":  "4:3",
		"768x1024":  "3:4",
		"1792x1024": "16:9",
		"1024x1792": "9:16",
		"1344x768":  "16:9",
		"768x1344":  "9:16",
	}

	if ratio, exists := sizeMap[size]; exists {
		return ratio
	}
	return "1:1" // default
}

// MarshalRequest marshals request to JSON
func MarshalRequest(request map[string]interface{}) ([]byte, error) {
	return json.Marshal(request)
}