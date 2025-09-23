package wavespeed

import "strings"

// Video Generation Models
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

// Image Generation Models
var ImageModels = []string{
	"wavespeed-ai/flux-1.1-pro",
	"wavespeed-ai/flux-dev",
	"wavespeed-ai/flux-kontext-pro/text-to-image",
	"wavespeed-ai/flux-kontext-dev",
	"wavespeed-ai/flux-kontext-max",
	"wavespeed-ai/flux-kontext-pro",
	"wavespeed-ai/flux-kontext-max/text-to-image",
	"google/gemini-2.5-flash-image/edit",
	"google/gemini-2.5-flash-image/text-to-image",
	"google/imagen3-fast",
	"google/imagen3",
	"google/imagen4-fast",
	"google/imagen4",
	"bytedance/seedream-v3",
	"bytedance/seedream-v4",
	"wavespeed-ai/step1x-edit",
}

// ModelList contains all supported models
var ModelList []string

func init() {
	ModelList = append(ModelList, VideoModels...)
	ModelList = append(ModelList, ImageModels...)
}

var ChannelName = "wavespeed"

// ModelType represents the type of generation model
type ModelType int

const (
	VideoGeneration ModelType = iota
	ImageGeneration
)

// ModelConfig contains configuration for a specific model
type ModelConfig struct {
	Type     ModelType
	Provider string
	Model    string
	Variant  string
	Endpoint string
}

// IsVideoModel checks if a model is for video generation
func IsVideoModel(modelPath string) bool {
	for _, vm := range VideoModels {
		if vm == modelPath {
			return true
		}
	}
	return false
}

// IsImageModel checks if a model is for image generation
func IsImageModel(modelPath string) bool {
	for _, im := range ImageModels {
		if im == modelPath {
			return true
		}
	}
	return false
}

// ParseModelPath extracts provider, model, and variant from model path
func ParseModelPath(modelPath string) (provider, model, variant string) {
	parts := strings.Split(modelPath, "/")
	if len(parts) >= 2 {
		provider = parts[0]
		if len(parts) == 2 {
			model = parts[1]
		} else if len(parts) >= 3 {
			model = parts[1]
			variant = strings.Join(parts[2:], "/")
		}
	}
	return
}

// GetModelConfig returns configuration for a given model
func GetModelConfig(modelPath string) *ModelConfig {
	provider, model, variant := ParseModelPath(modelPath)

	config := &ModelConfig{
		Provider: provider,
		Model:    model,
		Variant:  variant,
	}

	// Determine model type
	if IsVideoModel(modelPath) {
		config.Type = VideoGeneration
	} else {
		config.Type = ImageGeneration
	}

	// Build endpoint path
	if variant != "" {
		config.Endpoint = provider + "/" + model + "/" + variant
	} else {
		config.Endpoint = provider + "/" + model
	}

	return config
}

// IsValidModel checks if a model is supported
func IsValidModel(modelPath string) bool {
	for _, model := range ModelList {
		if model == modelPath {
			return true
		}
	}
	return false
}