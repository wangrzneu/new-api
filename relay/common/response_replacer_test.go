package common

import (
	"one-api/common"
	"strings"
	"testing"
)

func TestApplyLinkReplacement(t *testing.T) {
	// Setup test configuration
	originalEnabled := common.LinkReplacementEnabled
	originalDomain := common.ReplacementDomain
	defer func() {
		common.LinkReplacementEnabled = originalEnabled
		common.ReplacementDomain = originalDomain
	}()

	common.LinkReplacementEnabled = true
	common.ReplacementDomain = "proxy.example.com"

	tests := []struct {
		name     string
		input    string
		expected bool // 期望是否包含替换后的域名
	}{
		{
			name:     "JSON content",
			input:    `{"url": "https://example.com/api", "text": "Visit https://test.com"}`,
			expected: true,
		},
		{
			name:     "Plain text with markdown",
			input:    "Check out [OpenAI API](https://api.openai.com) for details",
			expected: true,
		},
		{
			name:     "Plain text with URL",
			input:    "Visit https://example.com for more info",
			expected: true,
		},
		{
			name:     "No links",
			input:    "This text has no links",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyLinkReplacement(tt.input)
			containsProxy := strings.Contains(result, "proxy.example.com")

			if tt.expected && !containsProxy {
				t.Errorf("Expected result to contain proxy domain, but got: %s", result)
			}
			if !tt.expected && containsProxy {
				t.Errorf("Expected result to not contain proxy domain, but got: %s", result)
			}
		})
	}
}

func TestApplyLinkReplacementToBytes(t *testing.T) {
	// Setup test configuration
	originalEnabled := common.LinkReplacementEnabled
	originalDomain := common.ReplacementDomain
	defer func() {
		common.LinkReplacementEnabled = originalEnabled
		common.ReplacementDomain = originalDomain
	}()

	common.LinkReplacementEnabled = true
	common.ReplacementDomain = "proxy.example.com"

	input := `{"message": "Visit https://example.com"}`
	inputBytes := []byte(input)

	result := ApplyLinkReplacementToBytes(inputBytes)
	resultStr := string(result)


	if !strings.Contains(resultStr, "proxy.example.com") {
		t.Errorf("Expected result to contain proxy domain, but got: %s", resultStr)
	}

	// 检查原始域名是否被替换，但排除代理域名本身
	if strings.Contains(resultStr, "https://example.com") {
		t.Errorf("Expected original URL to be replaced, but got: %s", resultStr)
	}
}

func TestApplyLinkReplacementDisabled(t *testing.T) {
	// Setup and restore configuration
	originalEnabled := common.LinkReplacementEnabled
	originalDomain := common.ReplacementDomain
	defer func() {
		common.LinkReplacementEnabled = originalEnabled
		common.ReplacementDomain = originalDomain
	}()

	// Disable link replacement
	common.LinkReplacementEnabled = false
	common.ReplacementDomain = "proxy.example.com"

	input := "Check out https://example.com"
	result := ApplyLinkReplacement(input)

	if result != input {
		t.Errorf("Expected no replacement when disabled, but got: %s", result)
	}
}