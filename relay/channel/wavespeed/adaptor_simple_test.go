package wavespeed

import (
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicFunctionality(t *testing.T) {
	// Test model classification
	t.Run("Model Classification", func(t *testing.T) {
		assert.True(t, IsVideoModel("minimax/hailuo-02/fast"))
		assert.False(t, IsVideoModel("wavespeed-ai/flux-1.1-pro"))
		assert.True(t, IsImageModel("wavespeed-ai/flux-1.1-pro"))
		assert.False(t, IsImageModel("minimax/hailuo-02/fast"))
		assert.True(t, IsValidModel("minimax/hailuo-02/fast"))
		assert.False(t, IsValidModel("invalid/model"))
	})

	// Test model configuration
	t.Run("Model Configuration", func(t *testing.T) {
		config := GetModelConfig("minimax/hailuo-02/fast")
		assert.Equal(t, "minimax", config.Provider)
		assert.Equal(t, "hailuo-02", config.Model)
		assert.Equal(t, "fast", config.Variant)
		assert.Equal(t, VideoGeneration, config.Type)
		assert.Equal(t, "minimax/hailuo-02/fast", config.Endpoint)

		config2 := GetModelConfig("wavespeed-ai/flux-1.1-pro")
		assert.Equal(t, "wavespeed-ai", config2.Provider)
		assert.Equal(t, "flux-1.1-pro", config2.Model)
		assert.Equal(t, "", config2.Variant)
		assert.Equal(t, ImageGeneration, config2.Type)
		assert.Equal(t, "wavespeed-ai/flux-1.1-pro", config2.Endpoint)
	})

	// Test model list
	t.Run("Model List", func(t *testing.T) {
		assert.Contains(t, ModelList, "minimax/hailuo-02/fast")
		assert.Contains(t, ModelList, "wavespeed-ai/flux-1.1-pro")
		assert.Greater(t, len(ModelList), 25) // Should have 30+ models
	})

	// Test endpoint generation
	t.Run("Endpoint Generation", func(t *testing.T) {
		endpoint := GetModelEndpoint("minimax/hailuo-02/fast")
		assert.Equal(t, "/api/v3/minimax/hailuo-02/fast", endpoint)

		endpoint2 := GetModelEndpoint("wavespeed-ai/flux-1.1-pro")
		assert.Equal(t, "/api/v3/wavespeed-ai/flux-1.1-pro", endpoint2)
	})
}

func TestAdaptorBasics(t *testing.T) {
	adaptor := &Adaptor{}

	t.Run("Channel Name", func(t *testing.T) {
		assert.Equal(t, "wavespeed", adaptor.GetChannelName())
	})

	t.Run("Model List", func(t *testing.T) {
		models := adaptor.GetModelList()
		assert.Greater(t, len(models), 25)
		assert.Contains(t, models, "minimax/hailuo-02/fast")
	})

	t.Run("Request URL Generation", func(t *testing.T) {
		info := &relaycommon.RelayInfo{
			ChannelMeta: &relaycommon.ChannelMeta{
				UpstreamModelName: "minimax/hailuo-02/fast",
				ChannelBaseUrl:    "https://api.wavespeed.ai",
			},
		}

		url, err := adaptor.GetRequestURL(info)
		assert.NoError(t, err)
		assert.Equal(t, "https://api.wavespeed.ai/api/v3/minimax/hailuo-02/fast", url)
	})

	t.Run("Invalid Model", func(t *testing.T) {
		info := &relaycommon.RelayInfo{
			ChannelMeta: &relaycommon.ChannelMeta{
				UpstreamModelName: "invalid/model",
				ChannelBaseUrl:    "https://api.wavespeed.ai",
			},
		}

		_, err := adaptor.GetRequestURL(info)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported model")
	})
}

func TestRequestConverterBasics(t *testing.T) {
	converter := NewRequestConverter()

	t.Run("Extract Prompt from Messages", func(t *testing.T) {
		messages := []dto.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there"},
			{Role: "user", Content: "Generate an image"},
		}

		prompt := converter.extractPromptFromMessages(messages)
		assert.Contains(t, prompt, "Hello")
		assert.Contains(t, prompt, "Hi there")
		assert.Contains(t, prompt, "Generate an image")
	})

	t.Run("Handle Quality Parameter", func(t *testing.T) {
		params := make(map[string]interface{})
		params["prompt"] = "test"

		result := converter.handleQualityParameter(params, "hd", "wavespeed-ai/flux-dev")
		assert.Equal(t, 50, result["num_inference_steps"])

		result2 := converter.handleQualityParameter(params, "standard", "wavespeed-ai/flux-dev")
		assert.Equal(t, 28, result2["num_inference_steps"])
	})

	t.Run("Handle Temperature Parameter", func(t *testing.T) {
		params := make(map[string]interface{})
		params["prompt"] = "test"

		result := converter.handleTemperatureParameter(params, 1.0, "wavespeed-ai/flux-dev")
		assert.Contains(t, result, "guidance_scale")

		guidanceScale, ok := result["guidance_scale"].(float64)
		assert.True(t, ok)
		assert.Greater(t, guidanceScale, 1.0)
	})
}

func TestResponseHandlerBasics(t *testing.T) {
	handler := NewResponseHandler()

	t.Run("Map Status to Finish Reason", func(t *testing.T) {
		assert.Equal(t, "stop", handler.mapStatusToFinishReason("completed"))
		assert.Equal(t, "stop", handler.mapStatusToFinishReason("failed"))
		assert.Equal(t, "length", handler.mapStatusToFinishReason("processing"))
		assert.Equal(t, "length", handler.mapStatusToFinishReason("created"))
	})

	t.Run("Estimate Tokens", func(t *testing.T) {
		resp := &WavespeedResponse{
			Outputs: []string{"url1", "url2"},
		}

		promptTokens := handler.estimatePromptTokens(resp)
		assert.Greater(t, promptTokens, 0)

		completionTokens := handler.estimateCompletionTokens(resp)
		assert.Greater(t, completionTokens, 0)
	})

	t.Run("Task Completion Check", func(t *testing.T) {
		assert.True(t, handler.IsTaskComplete("completed"))
		assert.True(t, handler.IsTaskComplete("failed"))
		assert.False(t, handler.IsTaskComplete("processing"))
		assert.False(t, handler.IsTaskComplete("created"))
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("Convert Size to Aspect Ratio", func(t *testing.T) {
		assert.Equal(t, "1:1", convertSizeToAspectRatio("1024x1024"))
		assert.Equal(t, "16:9", convertSizeToAspectRatio("1792x1024"))
		assert.Equal(t, "9:16", convertSizeToAspectRatio("1024x1792"))
		assert.Equal(t, "1:1", convertSizeToAspectRatio("unknown"))
	})

	t.Run("Aspect Ratio Model Check", func(t *testing.T) {
		assert.True(t, isAspectRatioModel("wavespeed-ai/flux-1.1-pro"))
		assert.True(t, isAspectRatioModel("google/imagen3"))
		assert.False(t, isAspectRatioModel("minimax/hailuo-02/fast"))
	})

	t.Run("Parse Model Path", func(t *testing.T) {
		provider, model, variant := ParseModelPath("minimax/hailuo-02/fast")
		assert.Equal(t, "minimax", provider)
		assert.Equal(t, "hailuo-02", model)
		assert.Equal(t, "fast", variant)

		provider2, model2, variant2 := ParseModelPath("wavespeed-ai/flux-1.1-pro")
		assert.Equal(t, "wavespeed-ai", provider2)
		assert.Equal(t, "flux-1.1-pro", model2)
		assert.Equal(t, "", variant2)
	})
}

// Performance tests
func BenchmarkParseModelPath(b *testing.B) {
	model := "minimax/hailuo-02/fast"
	for i := 0; i < b.N; i++ {
		ParseModelPath(model)
	}
}