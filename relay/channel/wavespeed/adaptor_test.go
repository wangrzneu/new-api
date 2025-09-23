package wavespeed

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdaptor_GetRequestURL(t *testing.T) {
	adaptor := &Adaptor{}

	tests := []struct {
		name      string
		model     string
		baseURL   string
		expected  string
		shouldErr bool
	}{
		{
			name:     "flux image model",
			model:    "wavespeed-ai/flux-1.1-pro",
			baseURL:  "https://api.wavespeed.ai",
			expected: "https://api.wavespeed.ai/api/v3/wavespeed-ai/flux-1.1-pro",
		},
		{
			name:     "hailuo video model",
			model:    "minimax/hailuo-02/fast",
			baseURL:  "https://api.wavespeed.ai",
			expected: "https://api.wavespeed.ai/api/v3/minimax/hailuo-02/fast",
		},
		{
			name:      "invalid model",
			model:     "invalid/model",
			baseURL:   "https://api.wavespeed.ai",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{
					UpstreamModelName: tt.model,
					ChannelBaseUrl:    tt.baseURL,
				},
			}

			url, err := adaptor.GetRequestURL(info)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, url)
			}
		})
	}
}

func TestAdaptor_ConvertImageRequest(t *testing.T) {
	adaptor := &Adaptor{}
	adaptor.Init(&relaycommon.RelayInfo{})

	tests := []struct {
		name     string
		model    string
		request  dto.ImageRequest
		hasError bool
	}{
		{
			name:  "basic image request",
			model: "wavespeed-ai/flux-1.1-pro",
			request: dto.ImageRequest{
				Prompt: "A beautiful sunset",
				Size:   "1024x1024",
				N:      1,
			},
		},
		{
			name:  "video request",
			model: "minimax/hailuo-02/fast",
			request: dto.ImageRequest{
				Prompt: "A person walking",
			},
		},
		{
			name:  "empty prompt",
			model: "wavespeed-ai/flux-1.1-pro",
			request: dto.ImageRequest{
				Size: "1024x1024",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{
					UpstreamModelName: tt.model,
				},
			}

			result, err := adaptor.ConvertImageRequest(c, info, tt.request)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify the result is a valid map
				resultMap, ok := result.(map[string]interface{})
				assert.True(t, ok)

				if tt.request.Prompt != "" {
					assert.Equal(t, tt.request.Prompt, resultMap["prompt"])
				}
			}
		})
	}
}

func TestAdaptor_ConvertOpenAIRequest(t *testing.T) {
	adaptor := &Adaptor{}
	adaptor.Init(&relaycommon.RelayInfo{})

	tests := []struct {
		name     string
		model    string
		request  *dto.GeneralOpenAIRequest
		hasError bool
	}{
		{
			name:  "text to image request",
			model: "wavespeed-ai/flux-1.1-pro",
			request: &dto.GeneralOpenAIRequest{
				Model: "wavespeed-ai/flux-1.1-pro",
				Messages: []dto.Message{
					{
						Role:    "user",
						Content: "Generate a beautiful landscape",
					},
				},
			},
		},
		{
			name:     "nil request",
			model:    "wavespeed-ai/flux-1.1-pro",
			request:  nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{
					UpstreamModelName: tt.model,
				},
			}

			result, err := adaptor.ConvertOpenAIRequest(c, info, tt.request)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestModelClassification(t *testing.T) {
	tests := []struct {
		model    string
		isVideo  bool
		isImage  bool
		isValid  bool
	}{
		{"minimax/hailuo-02/fast", true, false, true},
		{"wavespeed-ai/flux-1.1-pro", false, true, true},
		{"bytedance/seedream-v4", false, true, true},
		{"invalid/model", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			assert.Equal(t, tt.isVideo, IsVideoModel(tt.model))
			assert.Equal(t, tt.isImage, IsImageModel(tt.model))
			assert.Equal(t, tt.isValid, IsValidModel(tt.model))
		})
	}
}

func TestGetModelConfig(t *testing.T) {
	tests := []struct {
		model    string
		provider string
		modelName string
		variant  string
		endpoint string
		modelType ModelType
	}{
		{
			model:     "minimax/hailuo-02/fast",
			provider:  "minimax",
			modelName: "hailuo-02",
			variant:   "fast",
			endpoint:  "minimax/hailuo-02/fast",
			modelType: VideoGeneration,
		},
		{
			model:     "wavespeed-ai/flux-1.1-pro",
			provider:  "wavespeed-ai",
			modelName: "flux-1.1-pro",
			variant:   "",
			endpoint:  "wavespeed-ai/flux-1.1-pro",
			modelType: ImageGeneration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			config := GetModelConfig(tt.model)
			assert.Equal(t, tt.provider, config.Provider)
			assert.Equal(t, tt.modelName, config.Model)
			assert.Equal(t, tt.variant, config.Variant)
			assert.Equal(t, tt.endpoint, config.Endpoint)
			assert.Equal(t, tt.modelType, config.Type)
		})
	}
}

func TestMarshalRequest(t *testing.T) {
	request := map[string]interface{}{
		"prompt":       "test prompt",
		"aspect_ratio": "1:1",
	}

	data, err := MarshalRequest(request)
	assert.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, request["prompt"], unmarshaled["prompt"])
	assert.Equal(t, request["aspect_ratio"], unmarshaled["aspect_ratio"])
}

// Mock response for testing response handler
func createMockResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
}

func TestResponseHandler_HandleResponse(t *testing.T) {
	handler := NewResponseHandler()

	// Test successful response
	successBody := `{
		"id": "test-123",
		"status": "completed",
		"created_at": "2023-01-01T00:00:00Z",
		"model": "wavespeed-ai/flux-1.1-pro",
		"outputs": ["https://example.com/image1.jpg"],
		"has_nsfw_contents": [false]
	}`

	resp := &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}
	resp.Body = io.NopCloser(bytes.NewReader([]byte(successBody)))

	result, err := handler.HandleResponse(resp, "wavespeed-ai/flux-1.1-pro")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-123", result.Id)
	assert.Equal(t, "wavespeed-ai/flux-1.1-pro", result.Model)
}

// Benchmark tests
func BenchmarkGetModelConfig(b *testing.B) {
	model := "minimax/hailuo-02/fast"
	for i := 0; i < b.N; i++ {
		GetModelConfig(model)
	}
}

func BenchmarkIsVideoModel(b *testing.B) {
	model := "minimax/hailuo-02/fast"
	for i := 0; i < b.N; i++ {
		IsVideoModel(model)
	}
}