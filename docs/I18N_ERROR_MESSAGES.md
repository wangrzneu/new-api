# 错误消息国际化支持

本文档介绍了 one-api 项目中错误消息的国际化实现方案，支持中文和英文两种语言。

## 功能特性

- **双语言支持**: 支持中文（zh）和英文（en）
- **自动检测**: 根据 Accept-Language 头或查询参数自动选择语言
- **向后兼容**: 不影响现有错误处理逻辑
- **回退机制**: 翻译缺失时自动回退到英语或错误码

## 语言检测优先级

1. 查询参数 `?lang=zh` 或 `?lang=en`
2. HTTP 头 `Accept-Language: zh-CN,zh;q=0.9,en;q=0.8`
3. 默认英语

## 使用方法

### 1. 基本用法

```go
import "one-api/types"

// 创建国际化错误
err := types.NewI18nError(
    c, // gin.Context
    fmt.Errorf("some error"),
    types.ErrorCodeModelNotFound,
)

// 获取国际化错误消息
i18nMsg := err.ToI18nMessage(c)
fmt.Printf("错误: %s (语言: %s)", i18nMsg.Message, i18nMsg.Lang)
```

### 2. HTTP 响应中使用

```go
func handleRequest(c *gin.Context) {
    err := types.NewI18nError(c, someError, types.ErrorCodeInvalidRequest)
    i18nMsg := err.ToI18nMessage(c)

    c.JSON(http.StatusBadRequest, gin.H{
        "error": gin.H{
            "code":    i18nMsg.Code,
            "message": i18nMsg.Message,
            "lang":    i18nMsg.Lang,
        },
    })
}
```

### 3. OpenAI 格式错误

```go
err := types.NewI18nOpenAIError(
    c,
    fmt.Errorf("validation failed"),
    types.ErrorCodeInvalidRequest,
    http.StatusBadRequest,
)

openaiError := err.ToOpenAIError()
c.JSON(http.StatusBadRequest, gin.H{"error": openaiError})
```

## 响应示例

### 英文请求
```bash
curl -H "Accept-Language: en-US" http://localhost:8080/api/endpoint
```

响应:
```json
{
  "error": {
    "code": "model_not_found",
    "message": "Model not found",
    "lang": "en"
  }
}
```

### 中文请求
```bash
curl -H "Accept-Language: zh-CN" http://localhost:8080/api/endpoint
# 或
curl http://localhost:8080/api/endpoint?lang=zh
```

响应:
```json
{
  "error": {
    "code": "model_not_found",
    "message": "未找到模型",
    "lang": "zh"
  }
}
```

## 支持的错误码

目前支持以下错误码的国际化翻译：

| 错误码 | 英文 | 中文 |
|--------|------|------|
| `invalid_request` | Invalid request | 无效请求 |
| `do_request_failed` | Request failed | 请求失败 |
| `channel:invalid_key` | Invalid API key | 无效的API密钥 |
| `channel:response_time_exceeded` | Response time exceeded | 响应超时 |
| `model_not_found` | Model not found | 未找到模型 |
| `insufficient_user_quota` | Insufficient user quota | 用户配额不足 |
| `access_denied` | Access denied | 访问被拒绝 |
| `bad_response` | Bad response from upstream service | 上游服务响应异常 |
| 等等... | ... | ... |

完整列表请参考 `types/error_i18n.go` 中的 `ErrorMessages` 映射。

## 在 WaveSpeed 适配器中的应用

WaveSpeed 适配器已经更新为使用国际化错误消息：

```go
// 响应处理器中的错误处理
if wavespeedResp.Error != "" {
    return nil, types.NewI18nError(
        context.Background(),
        errors.New(wavespeedResp.Error),
        types.ErrorCodeDoRequestFailed,
    )
}
```

## 扩展新的翻译

要添加新的错误码翻译，请在 `types/error_i18n.go` 的 `ErrorMessages` 映射中添加：

```go
var ErrorMessages = map[ErrorCode]map[string]string{
    // 现有翻译...

    ErrorCodeNewError: {
        LangEN: "New error message",
        LangZH: "新的错误消息",
    },
}
```

## 测试

运行国际化相关测试：

```bash
go test ./types -v -run ".*I18n.*|.*Language.*|.*ErrorMessage.*"
```

## 最佳实践

1. **优先使用国际化错误**: 对于面向用户的错误，使用 `NewI18nError` 替代 `NewError`
2. **保持一致性**: 确保同一错误码在不同语言中表达相同的含义
3. **简洁明了**: 错误消息应该简洁明了，避免过于技术性的术语
4. **测试覆盖**: 为新添加的错误码编写测试用例