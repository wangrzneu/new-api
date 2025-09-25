package wavespeed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/service"

	"github.com/gin-gonic/gin"
)

// ImageModels 图像生成模型列表
var ImageModels = []string{
	"wavespeed-ai/flux-1.1-pro",
	"wavespeed-ai/flux-dev",
	"wavespeed-ai/flux-kontext-pro/text-to-image",
	"wavespeed-ai/flux-kontext-dev",
	"wavespeed-ai/flux-kontext-max",
	"wavespeed-ai/flux-kontext-pro",
	"wavespeed-ai/flux-kontext-max/text-to-image",
	"wavespeed-ai/step1x-edit",
	"google/gemini-2.5-flash-image/edit",
	"google/gemini-2.5-flash-image/text-to-image",
	"google/imagen3",
	"google/imagen3-fast",
	"google/imagen4",
	"google/imagen4-fast",
	"bytedance/seedream-v3",
	"bytedance/seedream-v4",
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

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	// 使用统一的图像生成 action
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionImageGenerate)
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if !isValidImageModel(info.UpstreamModelName) {
		return "", fmt.Errorf("unsupported image model: %s", info.UpstreamModelName)
	}

	endpoint := getImageModelEndpoint(info.UpstreamModelName)
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

	body, err := a.convertToImageRequestPayload(&req, info.Action)
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

type UnifiedResponse struct {
	TaskID string `json:"task_id"`
	Data   struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func convertToUnifiedResponse(respBody []byte) (*UnifiedResponse, error) {
	var wsResp WavespeedImageResponse
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

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}

	var wsResp WavespeedImageResponse
	err = json.Unmarshal(responseBody, &wsResp)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "unmarshal_response_failed", http.StatusInternalServerError)
		return
	}

	if wsResp.Data.Status != "created" && wsResp.Data.Status != "processing" {
		common.SysLog(fmt.Sprintf("wave raw response: %v", string(responseBody)))
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("task failed with status: %s", wsResp.Data.Status), "task_failed", http.StatusBadRequest)
		return
	}

	unifiedResponse := UnifiedResponse{
		TaskID: wsResp.Data.ID,
		Data: struct {
			TaskID string `json:"task_id"`
		}{
			TaskID: wsResp.Data.ID,
		},
		Code:    wsResp.Code,
		Message: wsResp.Message,
	}

	c.JSON(http.StatusOK, unifiedResponse)
	return wsResp.Data.ID, responseBody, nil
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

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	taskInfo := &relaycommon.TaskInfo{}
	resPayload := WavespeedImageResponse{}
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
			taskInfo.Url = strings.Join(resPayload.Data.Outputs, ",")
		}
	case "failed":
		taskInfo.Status = model.TaskStatusFailure
		taskInfo.Reason = "Task failed"
	default:
		return nil, fmt.Errorf("unknown task status: %s", resPayload.Data.Status)
	}

	return taskInfo, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return ImageModels
}

func (a *TaskAdaptor) GetChannelName() string {
	return "wavespeed_image"
}

// 数据结构定义
type WavespeedImageRequest struct {
	Prompt         string   `json:"prompt"`
	NegativePrompt string   `json:"negative_prompt,omitempty"`
	Image          string   `json:"image,omitempty"`
	Images         []string `json:"images,omitempty"`
	Mask           string   `json:"mask,omitempty"`
	Size           string   `json:"size,omitempty"`
	Style          string   `json:"style,omitempty"`
	Quality        string   `json:"quality,omitempty"`
	N              int      `json:"n,omitempty"`
	Seed           int64    `json:"seed,omitempty"`
	Steps          int      `json:"steps,omitempty"`
	CfgScale       float64  `json:"cfg_scale,omitempty"`
}

type WavespeedImageResponse struct {
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

func (a *TaskAdaptor) convertToImageRequestPayload(req *relaycommon.TaskSubmitReq, action string) (*WavespeedImageRequest, error) {
	payload := &WavespeedImageRequest{
		Prompt:   req.Prompt,
		Size:     a.convertImageSize(req.Size),
		N:        1,
		Steps:    30,
		CfgScale: 7.0,
		Seed:     -1,
	}

	// 如果有输入图像，设置相关参数（支持图像编辑、变体等功能）
	if req.Image != "" {
		payload.Image = req.Image
	}
	if len(req.Images) > 0 {
		payload.Images = req.Images
	}

	// 解析 metadata 中的其他参数
	if req.Metadata != nil {
		if negPrompt, ok := req.Metadata["negative_prompt"]; ok {
			payload.NegativePrompt = fmt.Sprintf("%v", negPrompt)
		}
		if mask, ok := req.Metadata["mask"]; ok {
			payload.Mask = fmt.Sprintf("%v", mask)
		}
		if n, ok := req.Metadata["n"]; ok {
			if num, err := strconv.Atoi(fmt.Sprintf("%v", n)); err == nil {
				payload.N = num
			}
		}
		if steps, ok := req.Metadata["steps"]; ok {
			if s, err := strconv.Atoi(fmt.Sprintf("%v", steps)); err == nil {
				payload.Steps = s
			}
		}
		if cfg, ok := req.Metadata["cfg_scale"]; ok {
			if c, err := strconv.ParseFloat(fmt.Sprintf("%v", cfg), 64); err == nil {
				payload.CfgScale = c
			}
		}
		if seed, ok := req.Metadata["seed"]; ok {
			if s, err := strconv.ParseInt(fmt.Sprintf("%v", seed), 10, 64); err == nil {
				payload.Seed = s
			}
		}
		if style, ok := req.Metadata["style"]; ok {
			payload.Style = fmt.Sprintf("%v", style)
		}
		if quality, ok := req.Metadata["quality"]; ok {
			payload.Quality = fmt.Sprintf("%v", quality)
		}
	}

	return payload, nil
}

func (a *TaskAdaptor) convertImageSize(size string) string {
	switch size {
	case "512x512":
		return "512*512"
	case "1024x1024":
		return "1024*1024"
	case "1024x768":
		return "1024*768"
	case "768x1024":
		return "768*1024"
	case "1280x720":
		return "1280*720"
	case "720x1280":
		return "720*1280"
	case "1920x1080":
		return "1920*1080"
	case "1080x1920":
		return "1080*1920"
	default:
		return "1024*1024"
	}
}

func isValidImageModel(modelPath string) bool {
	for _, model := range ImageModels {
		if model == modelPath {
			return true
		}
	}
	return false
}

func getImageModelEndpoint(modelPath string) string {
	return fmt.Sprintf("/api/v3/%s", modelPath)
}
