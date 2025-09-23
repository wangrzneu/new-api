package types

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		acceptLanguage string
		queryLang      string
		expected       string
	}{
		{
			name:           "English Accept-Language",
			acceptLanguage: "en-US,en;q=0.9",
			expected:       LangEN,
		},
		{
			name:           "Chinese Accept-Language",
			acceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
			expected:       LangZH,
		},
		{
			name:      "Query parameter zh",
			queryLang: "zh",
			expected:  LangZH,
		},
		{
			name:      "Query parameter en",
			queryLang: "en",
			expected:  LangEN,
		},
		{
			name:           "Query parameter overrides Accept-Language",
			acceptLanguage: "zh-CN",
			queryLang:      "en",
			expected:       LangEN,
		},
		{
			name:     "Default language when no headers",
			expected: LangEN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)

			// 设置请求
			req := &http.Request{
				Header: make(http.Header),
				URL:    &url.URL{},
			}

			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}

			if tt.queryLang != "" {
				values := url.Values{}
				values.Set("lang", tt.queryLang)
				req.URL.RawQuery = values.Encode()
			}

			c.Request = req

			result := GetLanguage(c)
			if result != tt.expected {
				t.Errorf("GetLanguage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetErrorMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		errorCode      ErrorCode
		acceptLanguage string
		expectedLang   string
		expectedMsg    string
	}{
		{
			name:           "Model not found - English",
			errorCode:      ErrorCodeModelNotFound,
			acceptLanguage: "en-US",
			expectedLang:   LangEN,
			expectedMsg:    "Model not found",
		},
		{
			name:           "Model not found - Chinese",
			errorCode:      ErrorCodeModelNotFound,
			acceptLanguage: "zh-CN",
			expectedLang:   LangZH,
			expectedMsg:    "未找到模型",
		},
		{
			name:           "Invalid request - English",
			errorCode:      ErrorCodeInvalidRequest,
			acceptLanguage: "en-US",
			expectedLang:   LangEN,
			expectedMsg:    "Invalid request",
		},
		{
			name:           "Invalid request - Chinese",
			errorCode:      ErrorCodeInvalidRequest,
			acceptLanguage: "zh-CN",
			expectedLang:   LangZH,
			expectedMsg:    "无效请求",
		},
		{
			name:           "Unknown error code - fallback to code string",
			errorCode:      "unknown_error_code",
			acceptLanguage: "zh-CN",
			expectedLang:   LangZH,
			expectedMsg:    "unknown_error_code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			req := &http.Request{
				Header: make(http.Header),
				URL:    &url.URL{},
			}
			req.Header.Set("Accept-Language", tt.acceptLanguage)
			c.Request = req

			result := GetErrorMessage(c, tt.errorCode)

			if result.Code != tt.errorCode {
				t.Errorf("GetErrorMessage().Code = %v, want %v", result.Code, tt.errorCode)
			}
			if result.Lang != tt.expectedLang {
				t.Errorf("GetErrorMessage().Lang = %v, want %v", result.Lang, tt.expectedLang)
			}
			if result.Message != tt.expectedMsg {
				t.Errorf("GetErrorMessage().Message = %v, want %v", result.Message, tt.expectedMsg)
			}
		})
	}
}

func TestNewI18nError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(nil)
	req := &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}
	req.Header.Set("Accept-Language", "zh-CN")
	c.Request = req

	err := NewI18nError(c, nil, ErrorCodeInvalidRequest)

	if err.GetErrorCode() != ErrorCodeInvalidRequest {
		t.Errorf("NewI18nError().GetErrorCode() = %v, want %v", err.GetErrorCode(), ErrorCodeInvalidRequest)
	}

	// 验证错误消息是否被国际化
	if err.Error() != "无效请求" {
		t.Errorf("NewI18nError().Error() = %v, want %v", err.Error(), "无效请求")
	}

	// 测试 ToI18nMessage 方法
	i18nMsg := err.ToI18nMessage(c)
	if i18nMsg.Lang != LangZH {
		t.Errorf("ToI18nMessage().Lang = %v, want %v", i18nMsg.Lang, LangZH)
	}
	if i18nMsg.Message != "无效请求" {
		t.Errorf("ToI18nMessage().Message = %v, want %v", i18nMsg.Message, "无效请求")
	}
}