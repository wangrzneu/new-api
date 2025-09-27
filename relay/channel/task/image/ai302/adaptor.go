package ai302

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
	// 302.ai 图片异步生成接口
	endpoint := "/302/v2/image/generate"
	// 添加 run_async=true 查询参数以启用异步模式
	return fmt.Sprintf("%s%s?run_async=true", a.baseURL, endpoint), nil
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
	if info.UpstreamModelName != "" {
		req.Model = info.UpstreamModelName
	}
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

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}

	var ai302Resp Ai302ImageResponse
	err = json.Unmarshal(responseBody, &ai302Resp)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "unmarshal_response_failed", http.StatusInternalServerError)
		return
	}

	// 302.ai 返回的状态应该是 pending 或 processing
	if ai302Resp.Status != "pending" && ai302Resp.Status != "processing" {
		common.SysLog(fmt.Sprintf("302.ai raw response: %v", string(responseBody)))
		taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("task failed with status: %s", ai302Resp.Status), "task_failed", http.StatusBadRequest)
		return
	}

	unifiedResponse := UnifiedResponse{
		TaskID: ai302Resp.TaskID,
		Data: struct {
			TaskID string `json:"task_id"`
		}{
			TaskID: ai302Resp.TaskID,
		},
		Code:    0, // 假设成功是 0
		Message: "success",
	}

	c.JSON(http.StatusOK, unifiedResponse)
	return ai302Resp.TaskID, responseBody, nil
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	// 302.ai 异步获取图片生成结果接口
	url := fmt.Sprintf("%s/302/v2/image/fetch/%s", baseUrl, taskID)
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
	resPayload := Ai302TaskResultResponse{}
	err := json.Unmarshal(respBody, &resPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	taskInfo.TaskID = resPayload.TaskID

	switch resPayload.Status {
	case "pending":
		taskInfo.Status = model.TaskStatusSubmitted
	case "processing":
		taskInfo.Status = model.TaskStatusInProgress
	case "completed":
		taskInfo.Status = model.TaskStatusSuccess
		// 处理生成的图片链接
		if resPayload.ImageURL != "" {
			taskInfo.Url = resPayload.ImageURL
		} else if len(resPayload.ImageURLs) > 0 {
			taskInfo.Url = strings.Join(resPayload.ImageURLs, ",")
		}
	case "failed":
		taskInfo.Status = model.TaskStatusFailure
		if resPayload.Error.Message != "" {
			taskInfo.Reason = resPayload.Error.Message
		} else {
			taskInfo.Reason = "Task failed"
		}
	default:
		return nil, fmt.Errorf("unknown task status: %s", resPayload.Status)
	}

	return taskInfo, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return []string{}
}

func (a *TaskAdaptor) GetChannelName() string {
	return "ai302_image"
}

// 302.ai 图片生成请求结构
type Ai302ImageRequest struct {
	Prompt         string      `json:"prompt"`
	Model          string      `json:"model"`
	Height         int         `json:"height,omitempty"`
	Width          int         `json:"width,omitempty"`
	NegativePrompt string      `json:"negative_prompt,omitempty"`
	AspectRatio    string      `json:"aspect_ratio,omitempty"`
	OutputFormat   string      `json:"output_format,omitempty"`
	Image          interface{} `json:"image,omitempty"` // 支持字符串或数组
	MaskImage      string      `json:"mask_image,omitempty"`
}

// 302.ai 图片生成响应结构 (提交任务)
type Ai302ImageResponse struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// 302.ai 任务结果响应结构 (获取结果)
type Ai302TaskResultResponse struct {
	Model       string   `json:"model"`
	TaskID      string   `json:"task_id"`
	UpdatedAt   string   `json:"updated_at"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"created_at"`
	CompletedAt string   `json:"completed_at"`
	ImageURL    string   `json:"image_url"`
	ImageURLs   []string `json:"image_urls"`
	Req         struct {
		Prompt string `json:"prompt"`
		Model  string `json:"model"`
	} `json:"req"`
	Error struct {
		Error     string `json:"error"`
		Message   string `json:"message"`
		ErrorType string `json:"error_type"`
	} `json:"error"`
}

func (a *TaskAdaptor) convertToImageRequestPayload(req *relaycommon.TaskSubmitReq, action string) (*Ai302ImageRequest, error) {
	payload := &Ai302ImageRequest{
		Prompt:       req.Prompt,
		Model:        req.Model,
		Height:       1024,
		Width:        1024,
		OutputFormat: "png",
	}

	// 处理图像尺寸
	if req.Size != "" {
		width, height := a.parseImageSize(req.Size)
		payload.Width = width
		payload.Height = height
		// 自动计算宽高比
		payload.AspectRatio = a.calculateAspectRatio(width, height)
	}

	// 如果有输入图像，设置相关参数
	if req.Image != "" {
		payload.Image = req.Image
	}
	if len(req.Images) > 0 {
		if len(req.Images) == 1 {
			payload.Image = req.Images[0]
		} else {
			payload.Image = req.Images // 支持多张图片输入
		}
	}

	// 解析 metadata 中的其他参数
	if req.Metadata != nil {
		if negPrompt, ok := req.Metadata["negative_prompt"]; ok {
			payload.NegativePrompt = fmt.Sprintf("%v", negPrompt)
		}
		if maskImage, ok := req.Metadata["mask_image"]; ok {
			payload.MaskImage = fmt.Sprintf("%v", maskImage)
		}
		if aspectRatio, ok := req.Metadata["aspect_ratio"]; ok {
			payload.AspectRatio = fmt.Sprintf("%v", aspectRatio)
		}
		if outputFormat, ok := req.Metadata["output_format"]; ok {
			payload.OutputFormat = fmt.Sprintf("%v", outputFormat)
		}
		if width, ok := req.Metadata["width"]; ok {
			if w, err := strconv.Atoi(fmt.Sprintf("%v", width)); err == nil {
				payload.Width = w
			}
		}
		if height, ok := req.Metadata["height"]; ok {
			if h, err := strconv.Atoi(fmt.Sprintf("%v", height)); err == nil {
				payload.Height = h
			}
		}
	}

	return payload, nil
}

func (a *TaskAdaptor) parseImageSize(size string) (int, int) {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 1024, 1024 // 默认尺寸
	}

	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return 1024, 1024 // 默认尺寸
	}

	return width, height
}

func (a *TaskAdaptor) calculateAspectRatio(width, height int) string {
	// 计算最大公约数
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}

	g := gcd(width, height)
	ratioW := width / g
	ratioH := height / g

	return fmt.Sprintf("%d:%d", ratioW, ratioH)
}
