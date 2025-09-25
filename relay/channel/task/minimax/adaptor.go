package minimax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"one-api/constant"
	"one-api/dto"
	"one-api/model"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/service"

	"github.com/pkg/errors"
)

// ============================
// Request / Response structures
// ============================

type requestPayload struct {
	Model           string `json:"model"`
	Prompt          string `json:"prompt,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	Resolution      string `json:"resolution,omitempty"`
	FirstFrameImage string `json:"first_frame_image,omitempty"`
	LastFrameImage  string `json:"last_frame_image,omitempty"`
}

type responsePayload struct {
	TaskId   string   `json:"task_id"`
	BaseResp baseResp `json:"base_resp"`
}

type baseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

type taskResultResponse struct {
	TaskId      string   `json:"task_id"`
	Status      string   `json:"status"`
	FileId      string   `json:"file_id,omitempty"`
	DownloadUrl string   `json:"download_url,omitempty"`
	BaseResp    baseResp `json:"base_resp"`
}

type fileRetrieveResponse struct {
	File     fileInfo `json:"file"`
	BaseResp baseResp `json:"base_resp"`
}

type fileInfo struct {
	FileId      int    `json:"file_id"`
	Bytes       int    `json:"bytes"`
	CreatedAt   int    `json:"created_at"`
	Filename    string `json:"filename"`
	Purpose     string `json:"purpose"`
	DownloadUrl string `json:"download_url"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	ChannelType int
	baseURL     string
	apiKey      string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
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

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return fmt.Sprintf("%s/v1/video_generation", a.baseURL), nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, _ *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}

	var mResp responsePayload
	err = json.Unmarshal(responseBody, &mResp)
	if err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrap(err, fmt.Sprintf("%s", responseBody)), "unmarshal_response_failed", http.StatusInternalServerError)
		return
	}

	if mResp.BaseResp.StatusCode != 0 {
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("task failed: %s", mResp.BaseResp.StatusMsg), "task_failed", http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, mResp)
	return mResp.TaskId, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	url := fmt.Sprintf("%s/v1/query/video_generation", baseUrl)

	reqBody := map[string]string{
		"task_id": taskID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	return service.GetHttpClient().Do(req)
}

func (a *TaskAdaptor) GetModelList() []string {
	return []string{"MiniMax-Hailuo-02"}
}

func (a *TaskAdaptor) GetChannelName() string {
	return "minimax"
}

// ============================
// helpers
// ============================

func (a *TaskAdaptor) convertToRequestPayload(req *relaycommon.TaskSubmitReq, modelName string) (*requestPayload, error) {
	r := requestPayload{
		Model:      modelName,
		Prompt:     req.Prompt,
		Duration:   defaultInt(req.Duration, 6),
		Resolution: defaultString(req.Size, "1080P"),
	}

	// Handle metadata to extract additional fields
	if req.Metadata != nil {
		metadata := req.Metadata
		if firstFrame, ok := metadata["first_frame_image"].(string); ok {
			r.FirstFrameImage = firstFrame
		}
		if lastFrame, ok := metadata["last_frame_image"].(string); ok {
			r.LastFrameImage = lastFrame
		}
		if duration, ok := metadata["duration"].(float64); ok && duration > 0 {
			r.Duration = int(duration)
		}
		if resolution, ok := metadata["resolution"].(string); ok && resolution != "" {
			r.Resolution = resolution
		}
	}

	return &r, nil
}

func defaultString(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func defaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// retrieveFileDownloadUrl 从 Minimax 文件检索 API 获取下载链接
func (a *TaskAdaptor) retrieveFileDownloadUrl(baseUrl, apiKey, fileId string) (string, error) {
	url := fmt.Sprintf("%s/v1/files/retrieve?file_id=%s", baseUrl, fileId)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := service.GetHttpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("file retrieve failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var fileResp fileRetrieveResponse
	err = json.Unmarshal(body, &fileResp)
	if err != nil {
		return "", err
	}

	if fileResp.BaseResp.StatusCode != 0 {
		return "", fmt.Errorf("file retrieve failed: %s", fileResp.BaseResp.StatusMsg)
	}

	return fileResp.File.DownloadUrl, nil
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	taskInfo := &relaycommon.TaskInfo{}

	var taskResp taskResultResponse
	err := json.Unmarshal(respBody, &taskResp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response body")
	}

	if taskResp.BaseResp.StatusCode != 0 {
		taskInfo.Status = model.TaskStatusFailure
		taskInfo.Reason = taskResp.BaseResp.StatusMsg
		return taskInfo, nil
	}

	switch taskResp.Status {
	case "submitted", "processing":
		taskInfo.Status = model.TaskStatusInProgress
	case "success":
		taskInfo.Status = model.TaskStatusSuccess
		if taskResp.DownloadUrl != "" {
			taskInfo.Url = taskResp.DownloadUrl
		} else if taskResp.FileId != "" {
			// 当任务成功但没有直接下载链接时，通过文件检索 API 获取下载链接
			if a.baseURL != "" && a.apiKey != "" {
				downloadUrl, err := a.retrieveFileDownloadUrl(a.baseURL, a.apiKey, taskResp.FileId)
				if err != nil {
					return nil, errors.Wrap(err, "failed to retrieve file download url")
				}
				taskInfo.Url = downloadUrl
			} else {
				return nil, errors.New("missing baseURL or apiKey for file retrieval")
			}
		}
	case "failed":
		taskInfo.Status = model.TaskStatusFailure
		taskInfo.Reason = "Generation failed"
	default:
		taskInfo.Status = model.TaskStatusInProgress
	}

	return taskInfo, nil
}
