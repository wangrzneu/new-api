package wavespeed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/service"

	"github.com/gin-gonic/gin"
)

// VideoModels contains all supported video generation models
var VideoModels = []string{
	"minimax/hailuo-02/fast",
	"minimax/hailuo-02/i2v-pro",
	"minimax/hailuo-02/standard",
	"kwaivgi/kling-v2.1-i2v-standard",
	"minimax/video-01",
	"bytedance/seedance-v1-pro-i2v-1080p",
	"bytedance/seedance-v1-pro-i2v-480p",
	"bytedance/seedance-v1-pro-i2v-720p",
	"wavespeed-ai/wan-2.1/i2v-480p",
	"wavespeed-ai/wan-2.1/i2v-480p-lora",
	"wavespeed-ai/wan-2.1/i2v-720p",
	"wavespeed-ai/wan-2.2/i2v-480p",
	"wavespeed-ai/wan-2.2/i2v-720p",
}

type TaskAdaptor struct {
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	endpoint := getModelEndpoint(info.UpstreamModelName)
	return fmt.Sprintf("%s%s", a.baseURL, endpoint), nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	v, exists := c.Get("task_request")
	if !exists {
		return nil, fmt.Errorf("request not found in context")
	}
	req := v.(relaycommon.TaskSubmitReq)

	body, err := a.convertToRequestPayload(&req, info.UpstreamModelName)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}

	unifiedResponse, err := convertToUnifiedResponse(responseBody)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "convert_to_unified_response_failed", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(unifiedResponse.Data)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "marshal_task_data_failed", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, unifiedResponse)
	return unifiedResponse.Data.TaskID, data, nil
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	url := fmt.Sprintf("%s/api/v3/predictions/%s/result", baseUrl, taskID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))

	return service.GetHttpClient().Do(req)
}

func (a *TaskAdaptor) GetModelList() []string {
	return VideoModels
}

func (a *TaskAdaptor) GetChannelName() string {
	return "wavespeed"
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	taskInfo := &relaycommon.TaskInfo{}
	resPayload := WavespeedResponse{}
	err := json.Unmarshal(respBody, &resPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	taskInfo.TaskID = resPayload.Data.ID

	switch resPayload.Data.Status {
	case "created":
		taskInfo.Status = model.TaskStatusSubmitted
	case "processing":
		taskInfo.Status = model.TaskStatusInProgress
	case "completed":
		taskInfo.Status = model.TaskStatusSuccess
		if len(resPayload.Data.Outputs) > 0 {
			taskInfo.Url = resPayload.Data.Outputs[0]
		}
	case "failed":
		taskInfo.Status = model.TaskStatusFailure
		taskInfo.Reason = "Task failed"
	default:
		return nil, fmt.Errorf("unknown task status: %s", resPayload.Data.Status)
	}

	return taskInfo, nil
}

type WavespeedRequest struct {
	Prompt            string  `json:"prompt"`
	Image             string  `json:"image,omitempty"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	Size              string  `json:"size,omitempty"`
	Duration          int     `json:"duration,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	GuidanceScale     float64 `json:"guidance_scale,omitempty"`
	FlowShift         float64 `json:"flow_shift,omitempty"`
	Seed              int     `json:"seed,omitempty"`
}

type WavespeedResponse struct {
	Data struct {
		ID              string            `json:"id"`
		Model           string            `json:"model"`
		Status          string            `json:"status"`
		CreatedAt       string            `json:"created_at"`
		Outputs         []string          `json:"outputs"`
		HasNsfwContents []bool            `json:"has_nsfw_contents"`
		URLs            map[string]string `json:"urls"`
	} `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type UnifiedResponse struct {
	TaskID string `json:"task_id"`
	Data   struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func convertToUnifiedResponse(respBody []byte) (*UnifiedResponse, error) {
	var wsResp WavespeedResponse
	err := json.Unmarshal(respBody, &wsResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	unifiedResp := UnifiedResponse{
		TaskID: wsResp.Data.ID,
		Data: struct {
			TaskID string `json:"task_id"`
		}{
			TaskID: wsResp.Data.ID,
		},
		Code:    wsResp.Code,
		Message: wsResp.Message,
	}

	return &unifiedResp, nil
}

func (a *TaskAdaptor) convertToRequestPayload(req *relaycommon.TaskSubmitReq, modelName string) (*WavespeedRequest, error) {
	payload := &WavespeedRequest{
		Prompt:   req.Prompt,
		Image:    req.Image,
		Size:     req.Size,
		Duration: req.Duration,
	}

	// Parse metadata for additional parameters
	if req.Metadata != nil {
		if duration, ok := req.Metadata["duration"]; ok {
			if d, err := strconv.Atoi(fmt.Sprintf("%v", duration)); err == nil {
				payload.Duration = d
			}
		}
		if negativePrompt, ok := req.Metadata["negative_prompt"]; ok {
			payload.NegativePrompt = fmt.Sprintf("%v", negativePrompt)
		}
		if steps, ok := req.Metadata["num_inference_steps"]; ok {
			if s, err := strconv.Atoi(fmt.Sprintf("%v", steps)); err == nil {
				payload.NumInferenceSteps = s
			}
		}
		if guidance, ok := req.Metadata["guidance_scale"]; ok {
			if g, err := strconv.ParseFloat(fmt.Sprintf("%v", guidance), 64); err == nil {
				payload.GuidanceScale = g
			}
		}
		if flowShift, ok := req.Metadata["flow_shift"]; ok {
			if f, err := strconv.ParseFloat(fmt.Sprintf("%v", flowShift), 64); err == nil {
				payload.FlowShift = f
			}
		}
		if seed, ok := req.Metadata["seed"]; ok {
			if s, err := strconv.Atoi(fmt.Sprintf("%v", seed)); err == nil {
				payload.Seed = s
			}
		}
	}

	return payload, nil
}

// Helper functions for model validation and endpoint generation
func isValidModel(modelPath string) bool {
	for _, model := range VideoModels {
		if model == modelPath {
			return true
		}
	}
	return false
}

func getModelEndpoint(modelPath string) string {
	// Convert model path to API endpoint
	// e.g., "wavespeed-ai/wan-2.1/i2v-480p" -> "/api/v3/wavespeed-ai/wan-2.1/i2v-480p"
	return fmt.Sprintf("/api/v3/%s", modelPath)
}
