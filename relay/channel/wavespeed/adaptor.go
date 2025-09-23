package wavespeed

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	relaycommon "one-api/relay/common"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
	requestConverter  *RequestConverter
	responseHandler   *ResponseHandler
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	if a.requestConverter == nil {
		a.requestConverter = NewRequestConverter()
	}
	if a.responseHandler == nil {
		a.responseHandler = NewResponseHandler()
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if !IsValidModel(info.UpstreamModelName) {
		return "", fmt.Errorf("unsupported model: %s", info.UpstreamModelName)
	}

	endpoint := GetModelEndpoint(info.UpstreamModelName)
	return fmt.Sprintf("%s%s", info.ChannelBaseUrl, endpoint), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	req.Set("Authorization", fmt.Sprintf("Bearer %s", info.ApiKey))
	req.Set("Content-Type", "application/json")
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	if a.requestConverter == nil {
		a.requestConverter = NewRequestConverter()
	}

	return a.requestConverter.ConvertOpenAIRequest(request, info.UpstreamModelName)
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	if a.requestConverter == nil {
		a.requestConverter = NewRequestConverter()
	}

	modelPath := info.UpstreamModelName
	if IsVideoModel(modelPath) {
		return a.requestConverter.ConvertVideoRequest(request, modelPath)
	} else {
		return a.requestConverter.ConvertImageRequest(request, modelPath)
	}
}

func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
	return nil, errors.New("Gemini requests not supported for WaveSpeed")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return nil, errors.New("Claude requests not supported for WaveSpeed")
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("audio requests not supported for WaveSpeed")
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, errors.New("rerank requests not supported for WaveSpeed")
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return nil, errors.New("embedding requests not supported for WaveSpeed")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return nil, errors.New("OpenAI responses requests not supported for WaveSpeed")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	if a.responseHandler == nil {
		a.responseHandler = NewResponseHandler()
	}

	// Handle the response using the response handler
	openaiResponse, apiErr := a.responseHandler.HandleResponse(resp, info.UpstreamModelName)
	if apiErr != nil {
		return nil, apiErr
	}

	// Send the response to the client
	c.JSON(http.StatusOK, openaiResponse)
	return &openaiResponse.Usage, nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

// Helper method to convert request to bytes
func (a *Adaptor) requestToBytes(request interface{}) (io.Reader, error) {
	// First try to convert to map[string]interface{}
	var requestMap map[string]interface{}
	if reqMap, ok := request.(map[string]interface{}); ok {
		requestMap = reqMap
	} else {
		return nil, fmt.Errorf("request must be of type map[string]interface{}")
	}

	requestBytes, err := MarshalRequest(requestMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	return bytes.NewReader(requestBytes), nil
}