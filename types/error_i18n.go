package types

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	LangEN = "en" // 英语（默认）
	LangZH = "zh" // 中文
)

type ErrorMessage struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Lang    string    `json:"lang"`
}

// 通用错误消息映射
var ErrorMessages = map[ErrorCode]map[string]string{
	ErrorCodeInvalidRequest: {
		LangEN: "Invalid request",
		LangZH: "无效请求",
	},
	ErrorCodeDoRequestFailed: {
		LangEN: "Request failed",
		LangZH: "请求失败",
	},
	ErrorCodeChannelInvalidKey: {
		LangEN: "Invalid API key",
		LangZH: "无效的API密钥",
	},
	ErrorCodeChannelResponseTimeExceeded: {
		LangEN: "Response time exceeded",
		LangZH: "响应超时",
	},
	ErrorCodeModelNotFound: {
		LangEN: "Model not found",
		LangZH: "未找到模型",
	},
	ErrorCodeInsufficientUserQuota: {
		LangEN: "Insufficient user quota",
		LangZH: "用户配额不足",
	},
	ErrorCodeAccessDenied: {
		LangEN: "Access denied",
		LangZH: "访问被拒绝",
	},
	ErrorCodeBadResponse: {
		LangEN: "Bad response from upstream service",
		LangZH: "上游服务响应异常",
	},
	ErrorCodeReadResponseBodyFailed: {
		LangEN: "Failed to read response body",
		LangZH: "读取响应内容失败",
	},
	ErrorCodeConvertRequestFailed: {
		LangEN: "Failed to convert request format",
		LangZH: "请求格式转换失败",
	},
	ErrorCodeCountTokenFailed: {
		LangEN: "Failed to count tokens",
		LangZH: "计算令牌数量失败",
	},
	ErrorCodeJsonMarshalFailed: {
		LangEN: "Failed to marshal JSON",
		LangZH: "JSON序列化失败",
	},
	ErrorCodeGetChannelFailed: {
		LangEN: "Failed to get channel",
		LangZH: "获取频道失败",
	},
	ErrorCodeBadResponseStatusCode: {
		LangEN: "Bad response status code",
		LangZH: "响应状态码异常",
	},
	ErrorCodeEmptyResponse: {
		LangEN: "Empty response",
		LangZH: "响应为空",
	},
	ErrorCodeSensitiveWordsDetected: {
		LangEN: "Sensitive words detected",
		LangZH: "检测到敏感词汇",
	},
	ErrorCodeModelPriceError: {
		LangEN: "Model price error",
		LangZH: "模型价格错误",
	},
	ErrorCodeInvalidApiType: {
		LangEN: "Invalid API type",
		LangZH: "无效的API类型",
	},
	ErrorCodeGenRelayInfoFailed: {
		LangEN: "Failed to generate relay info",
		LangZH: "生成中继信息失败",
	},
	ErrorCodeChannelNoAvailableKey: {
		LangEN: "No available key",
		LangZH: "无可用密钥",
	},
	ErrorCodeChannelParamOverrideInvalid: {
		LangEN: "Invalid parameter override",
		LangZH: "无效的参数覆盖",
	},
	ErrorCodeChannelHeaderOverrideInvalid: {
		LangEN: "Invalid header override",
		LangZH: "无效的请求头覆盖",
	},
	ErrorCodeChannelModelMappedError: {
		LangEN: "Model mapping error",
		LangZH: "模型映射错误",
	},
	ErrorCodeChannelAwsClientError: {
		LangEN: "AWS client error",
		LangZH: "AWS客户端错误",
	},
	ErrorCodeReadRequestBodyFailed: {
		LangEN: "Failed to read request body",
		LangZH: "读取请求内容失败",
	},
	ErrorCodeBadResponseBody: {
		LangEN: "Bad response body",
		LangZH: "响应内容异常",
	},
	ErrorCodeAwsInvokeError: {
		LangEN: "AWS invoke error",
		LangZH: "AWS调用错误",
	},
	ErrorCodeQueryDataError: {
		LangEN: "Query data error",
		LangZH: "查询数据错误",
	},
	ErrorCodeUpdateDataError: {
		LangEN: "Update data error",
		LangZH: "更新数据错误",
	},
	ErrorCodePreConsumeTokenQuotaFailed: {
		LangEN: "Failed to pre-consume token quota",
		LangZH: "预消费令牌配额失败",
	},
}

// 获取语言偏好
func GetLanguage(ctx context.Context) string {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		// 1. 查询参数优先
		if lang := ginCtx.Query("lang"); lang != "" {
			return normalizeLanguage(lang)
		}

		// 2. Accept-Language 头
		if acceptLang := ginCtx.GetHeader("Accept-Language"); acceptLang != "" {
			return parseAcceptLanguage(acceptLang)
		}
	}

	return LangEN // 默认英语
}

// 标准化语言代码
func normalizeLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if strings.HasPrefix(lang, "zh") {
		return LangZH
	}
	return LangEN
}

// 解析 Accept-Language 头
func parseAcceptLanguage(acceptLang string) string {
	languages := strings.Split(acceptLang, ",")
	for _, lang := range languages {
		if idx := strings.Index(lang, ";"); idx > 0 {
			lang = lang[:idx]
		}
		lang = strings.TrimSpace(lang)
		if strings.HasPrefix(strings.ToLower(lang), "zh") {
			return LangZH
		}
	}
	return LangEN
}

// 获取错误消息
func GetErrorMessage(ctx context.Context, code ErrorCode) ErrorMessage {
	lang := GetLanguage(ctx)

	// 查找翻译
	if messages, exists := ErrorMessages[code]; exists {
		if message, exists := messages[lang]; exists {
			return ErrorMessage{
				Code:    code,
				Message: message,
				Lang:    lang,
			}
		}
		// 回退到英语
		if message, exists := messages[LangEN]; exists {
			return ErrorMessage{
				Code:    code,
				Message: message,
				Lang:    LangEN,
			}
		}
	}

	// 最终回退到错误码
	return ErrorMessage{
		Code:    code,
		Message: string(code),
		Lang:    lang,
	}
}

// 为 NewAPIError 添加国际化方法
func (e *NewAPIError) ToI18nMessage(ctx context.Context) ErrorMessage {
	return GetErrorMessage(ctx, e.errorCode)
}

// 创建国际化错误的便捷函数
func NewI18nError(ctx context.Context, err error, errorCode ErrorCode, ops ...NewAPIErrorOptions) *NewAPIError {
	apiErr := NewError(err, errorCode, ops...)

	// 使用国际化消息更新错误文本
	i18nMsg := GetErrorMessage(ctx, errorCode)
	if i18nMsg.Message != string(errorCode) {
		apiErr.SetMessage(i18nMsg.Message)
	}

	return apiErr
}

// 为现有的错误创建函数添加国际化版本
func NewI18nOpenAIError(ctx context.Context, err error, errorCode ErrorCode, statusCode int, ops ...NewAPIErrorOptions) *NewAPIError {
	apiErr := NewOpenAIError(err, errorCode, statusCode, ops...)

	// 使用国际化消息更新错误文本
	i18nMsg := GetErrorMessage(ctx, errorCode)
	if i18nMsg.Message != string(errorCode) {
		apiErr.SetMessage(i18nMsg.Message)
		// 同时更新 RelayError 中的消息
		if openaiError, ok := apiErr.RelayError.(OpenAIError); ok {
			openaiError.Message = i18nMsg.Message
			apiErr.RelayError = openaiError
		}
	}

	return apiErr
}

func NewI18nErrorWithStatusCode(ctx context.Context, err error, errorCode ErrorCode, statusCode int, ops ...NewAPIErrorOptions) *NewAPIError {
	apiErr := NewErrorWithStatusCode(err, errorCode, statusCode, ops...)

	// 使用国际化消息更新错误文本
	i18nMsg := GetErrorMessage(ctx, errorCode)
	if i18nMsg.Message != string(errorCode) {
		apiErr.SetMessage(i18nMsg.Message)
		// 同时更新 RelayError 中的消息
		if openaiError, ok := apiErr.RelayError.(OpenAIError); ok {
			openaiError.Message = i18nMsg.Message
			apiErr.RelayError = openaiError
		}
	}

	return apiErr
}