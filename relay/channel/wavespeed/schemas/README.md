# WaveSpeed.ai API Schemas

This directory contains the complete OpenAPI 3.0 schemas for WaveSpeed.ai models, organized by model type with each model in its own dedicated file.

## Directory Structure

```
schemas/
├── video/                          # Video generation models
│   ├── bytedance-seedance-v1-pro-i2v-1080p.json
│   ├── bytedance-seedance-v1-pro-i2v-480p.json
│   ├── bytedance-seedance-v1-pro-i2v-720p.json
│   ├── kwaivgi-kling-v2.1-i2v-standard.json
│   ├── minimax-hailuo-02-fast.json
│   ├── minimax-hailuo-02-i2v-pro.json
│   ├── minimax-hailuo-02-standard.json
│   ├── minimax-video-01.json
│   ├── wavespeed-ai-wan-2.1-i2v-480p-lora.json
│   ├── wavespeed-ai-wan-2.1-i2v-480p.json
│   ├── wavespeed-ai-wan-2.1-i2v-720p.json
│   ├── wavespeed-ai-wan-2.2-i2v-480p.json
│   └── wavespeed-ai-wan-2.2-i2v-720p.json
├── image/                          # Image generation models
│   ├── bytedance-seedream-v3.json
│   ├── bytedance-seedream-v4.json
│   ├── google-gemini-2.5-flash-image-edit.json
│   ├── google-gemini-2.5-flash-image-text-to-image.json
│   ├── google-imagen3-fast.json
│   ├── google-imagen3.json
│   ├── google-imagen4-fast.json
│   ├── google-imagen4.json
│   ├── wavespeed-ai-flux-1.1-pro.json
│   ├── wavespeed-ai-flux-dev.json
│   ├── wavespeed-ai-flux-kontext-dev.json
│   ├── wavespeed-ai-flux-kontext-max-text-to-image.json
│   ├── wavespeed-ai-flux-kontext-max.json
│   ├── wavespeed-ai-flux-kontext-pro-text-to-image.json
│   ├── wavespeed-ai-flux-kontext-pro.json
│   └── wavespeed-ai-step1x-edit.json
└── README.md                       # This file
```

## API Overview

### Base URL
```
https://api.wavespeed.ai
```

### Authentication
All requests require an API key in the Authorization header:
```http
Authorization: Bearer YOUR_API_KEY
```

### Common Endpoints Pattern

1. **Create Prediction**
   ```
   POST /api/v3/{provider}/{model}/{variant}
   ```

2. **Get Result**
   ```
   GET /api/v3/predictions/{request_id}/result
   ```

## Video Generation Models

| Model | File | Description |
|-------|------|-------------|
| `minimax/hailuo-02/fast` | `video/minimax-hailuo-02-fast.json` | Fast image-to-video conversion |
| `minimax/hailuo-02/i2v-pro` | `video/minimax-hailuo-02-i2v-pro.json` | Professional image-to-video conversion |
| `minimax/hailuo-02/standard` | `video/minimax-hailuo-02-standard.json` | Balanced image-to-video conversion |
| `kwaivgi/kling-v2.1-i2v-standard` | `video/kwaivgi-kling-v2.1-i2v-standard.json` | Kling v2.1 standard image-to-video |
| `minimax/video-01` | `video/minimax-video-01.json` | Minimax video generation |
| `bytedance/seedance-v1-pro-i2v-1080p` | `video/bytedance-seedance-v1-pro-i2v-1080p.json` | Seedance V1 Pro 1080p image-to-video |
| `bytedance/seedance-v1-pro-i2v-480p` | `video/bytedance-seedance-v1-pro-i2v-480p.json` | Seedance V1 Pro 480p image-to-video |
| `bytedance/seedance-v1-pro-i2v-720p` | `video/bytedance-seedance-v1-pro-i2v-720p.json` | Seedance V1 Pro 720p image-to-video |
| `wavespeed-ai/wan-2.1/i2v-480p` | `video/wavespeed-ai-wan-2.1-i2v-480p.json` | WAN 2.1 480p video generation |
| `wavespeed-ai/wan-2.1/i2v-480p-lora` | `video/wavespeed-ai-wan-2.1-i2v-480p-lora.json` | WAN 2.1 480p LoRA video generation |
| `wavespeed-ai/wan-2.1/i2v-720p` | `video/wavespeed-ai-wan-2.1-i2v-720p.json` | WAN 2.1 720p video generation |
| `wavespeed-ai/wan-2.2/i2v-480p` | `video/wavespeed-ai-wan-2.2-i2v-480p.json` | WAN 2.2 480p video generation |
| `wavespeed-ai/wan-2.2/i2v-720p` | `video/wavespeed-ai-wan-2.2-i2v-720p.json` | WAN 2.2 720p video generation |

### Common Video Parameters
- `prompt` (string) - Video description
- `image` (string) - Input image URL or base64
- `duration` (integer) - Video duration in seconds
- `enable_prompt_expansion` (boolean) - Enhance prompts automatically

## Image Generation Models

| Model | File | Description |
|-------|------|-------------|
| `wavespeed-ai/flux-1.1-pro` | `image/wavespeed-ai-flux-1.1-pro.json` | High-quality text-to-image |
| `wavespeed-ai/flux-dev` | `image/wavespeed-ai-flux-dev.json` | FLUX development text-to-image |
| `wavespeed-ai/flux-kontext-pro/text-to-image` | `image/wavespeed-ai-flux-kontext-pro-text-to-image.json` | FLUX Kontext Pro text-to-image |
| `wavespeed-ai/flux-kontext-dev` | `image/wavespeed-ai-flux-kontext-dev.json` | FLUX Kontext development model |
| `wavespeed-ai/flux-kontext-max` | `image/wavespeed-ai-flux-kontext-max.json` | FLUX Kontext Max image generation |
| `wavespeed-ai/flux-kontext-pro` | `image/wavespeed-ai-flux-kontext-pro.json` | FLUX Kontext Pro image generation |
| `wavespeed-ai/flux-kontext-max/text-to-image` | `image/wavespeed-ai-flux-kontext-max-text-to-image.json` | FLUX Kontext Max text-to-image |
| `google/gemini-2.5-flash-image/edit` | `image/google-gemini-2.5-flash-image-edit.json` | Gemini 2.5 Flash image editing |
| `google/gemini-2.5-flash-image/text-to-image` | `image/google-gemini-2.5-flash-image-text-to-image.json` | Gemini 2.5 Flash text-to-image |
| `google/imagen3-fast` | `image/google-imagen3-fast.json` | Imagen 3 fast text-to-image |
| `google/imagen3` | `image/google-imagen3.json` | Imagen 3 text-to-image |
| `google/imagen4-fast` | `image/google-imagen4-fast.json` | Imagen 4 fast text-to-image |
| `google/imagen4` | `image/google-imagen4.json` | Imagen 4 text-to-image |
| `bytedance/seedream-v3` | `image/bytedance-seedream-v3.json` | Seedream V3 text-to-image |
| `bytedance/seedream-v4` | `image/bytedance-seedream-v4.json` | Seedream V4 text-to-image |
| `wavespeed-ai/step1x-edit` | `image/wavespeed-ai-step1x-edit.json` | Step1X image editing |

### Common Image Parameters
- `prompt` (string, required) - Image description
- `aspect_ratio` (string) - Image aspect ratio (1:1, 16:9, 9:16, 4:3, 3:4)
- `output_format` (string) - Output format (jpg, png)
- `seed` (integer) - Random seed for reproducibility

## Response Format

All models return a standard response format:

```json
{
  "id": "string",
  "status": "created|processing|completed|failed",
  "created_at": "ISO_timestamp",
  "model": "model_identifier",
  "outputs": ["url1", "url2"],
  "has_nsfw_contents": [false, false]
}
```

## Usage Examples

### Video Generation Example
```bash
curl -X POST "https://api.wavespeed.ai/api/v3/minimax/hailuo-02/fast" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "image": "https://example.com/image.jpg",
    "prompt": "A person walking in the park",
    "duration": 6
  }'
```

### Image Generation Example
```bash
curl -X POST "https://api.wavespeed.ai/api/v3/wavespeed-ai/flux-1.1-pro" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "A beautiful sunset over mountains",
    "aspect_ratio": "16:9"
  }'
```

### Get Results Example
```bash
curl -X GET "https://api.wavespeed.ai/api/v3/predictions/{request_id}/result" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## OpenAPI Specification

Each JSON file contains a complete OpenAPI 3.0 specification including:

- **Info**: Model name, description, and version
- **Servers**: API base URL
- **Paths**: Available endpoints with request/response schemas
- **Components**:
  - Input schemas with parameter definitions
  - Response schemas
  - Security schemes (API key authentication)
- **Security**: Authentication requirements

## Model-Specific Notes

### Minimax Hailuo 02 Series
- **Fast**: Optimized for speed with quality trade-offs
- **I2V Pro**: Professional quality with end_image support
- **Standard**: Balanced quality and speed

### WaveSpeed WAN Series
- **2.1**: Supports various resolutions and inference steps
- **2.2**: Enhanced with last_image parameter for better control

### Google Models
- **Gemini 2.5 Flash**: Fast generation with simple parameters
- **Imagen 3**: Advanced features with aspect ratio control

## Schema Sources

These schemas were extracted from the WaveSpeed.ai API documentation:
```
https://wavespeed.ai/center/default/api/v1/model_schema/{model_path}
```

For the most up-to-date schemas, you can fetch them directly using this pattern.

## Integration

Each schema file can be used independently for:
- API client generation
- Request validation
- Documentation generation
- Testing frameworks

Simply import the relevant JSON file for your target model to get complete API specifications.