package common

import (
	"strings"
	"testing"
)

func TestReplaceLinksInText(t *testing.T) {
	// Setup test configuration
	LinkReplacementEnabled = true
	ReplacementDomain = "proxy.example.com"

	tests := []struct {
		name     string
		input    string
		expected bool // 期望是否包含替换后的域名
	}{
		{
			name:     "Simple HTTP link",
			input:    "Check out this link: http://example.com/page",
			expected: true,
		},
		{
			name:     "HTTPS link in text",
			input:    "Visit https://api.openai.com/v1/models for more info",
			expected: true,
		},
		{
			name:     "Multiple links",
			input:    "Links: https://example1.com and http://example2.com",
			expected: true,
		},
		{
			name:     "Markdown link",
			input:    "Check out [OpenAI API](https://api.openai.com/v1/models) for details",
			expected: true,
		},
		{
			name:     "Multiple markdown links",
			input:    "Visit [Site 1](https://example1.com) and [Site 2](https://example2.com)",
			expected: true,
		},
		{
			name:     "Mixed markdown and plain links",
			input:    "See [API docs](https://api.example.com) or visit https://example.com directly",
			expected: true,
		},
		{
			name:     "No links",
			input:    "This text has no links at all",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceLinksInText(tt.input)
			containsProxy := contains(result, "proxy.example.com")

			if tt.expected && !containsProxy {
				t.Errorf("Expected result to contain proxy domain, but got: %s", result)
			}
			if !tt.expected && containsProxy {
				t.Errorf("Expected result to not contain proxy domain, but got: %s", result)
			}
		})
	}
}

func TestReplaceLinksInJSON(t *testing.T) {
	// Setup test configuration
	LinkReplacementEnabled = true
	ReplacementDomain = "proxy.example.com"

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "JSON with URL in string value",
			input:    `{"url": "https://example.com/api", "text": "Visit https://test.com"}`,
			expected: true,
		},
		{
			name:     "JSON with Markdown link",
			input:    `{"message": "Check [OpenAI](https://api.openai.com) for details"}`,
			expected: true,
		},
		{
			name:     "JSON with mixed plain and markdown links",
			input:    `{"content": "Visit https://example.com or see [docs](https://docs.example.com)"}`,
			expected: true,
		},
		{
			name:     "JSON without URLs",
			input:    `{"name": "test", "value": 123}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceLinksInJSON(tt.input)
			containsProxy := contains(result, "proxy.example.com")

			if tt.expected && !containsProxy {
				t.Errorf("Expected result to contain proxy domain, but got: %s", result)
			}
			if !tt.expected && containsProxy {
				t.Errorf("Expected result to not contain proxy domain, but got: %s", result)
			}
		})
	}
}

func TestDecodeLink(t *testing.T) {
	// Setup test configuration
	originalEnabled := LinkReplacementEnabled
	originalDomain := ReplacementDomain
	defer func() {
		LinkReplacementEnabled = originalEnabled
		ReplacementDomain = originalDomain
	}()

	LinkReplacementEnabled = true
	ReplacementDomain = "proxy.example.com"

	original := "https://example.com/api"
	encoded := encodeLink(original)

	// Extract the encoded part from the proxy URL
	// Format: https://proxy.example.com/redirect/{encoded}
	parts := strings.Split(encoded, "/redirect/")
	if len(parts) != 2 {
		t.Fatalf("Invalid encoded URL format: %s", encoded)
	}

	decoded, err := DecodeLink(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode link: %v", err)
	}

	if decoded != original {
		t.Errorf("Expected decoded link to be %s, but got %s", original, decoded)
	}

	// 验证编码后的字符串看起来像随机十六进制
	encodedPart := parts[1]
	if len(encodedPart) < 32 { // AES加密后应该至少32个字符
		t.Errorf("Encoded string too short, might be easily identifiable: %s", encodedPart)
	}

	// 验证字符串只包含十六进制字符
	for _, char := range encodedPart {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Encoded string contains non-hex character: %c in %s", char, encodedPart)
		}
	}
}

func TestDisabledLinkReplacement(t *testing.T) {
	// Disable link replacement
	LinkReplacementEnabled = false

	input := "Check out https://example.com"
	result := ReplaceLinksInText(input)

	if result != input {
		t.Errorf("Expected no replacement when disabled, but got: %s", result)
	}
}

func TestMarkdownLinkReplacement(t *testing.T) {
	// Setup test configuration
	LinkReplacementEnabled = true
	ReplacementDomain = "proxy.example.com"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "Simple markdown link",
			input: "[Google](https://google.com)",
		},
		{
			name:  "Markdown link with complex text",
			input: "[OpenAI API Documentation](https://api.openai.com/v1/models)",
		},
		{
			name:  "Multiple markdown links",
			input: "Check [Site 1](https://example1.com) and [Site 2](https://example2.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceLinksInText(tt.input)

			// 验证结果包含代理域名
			if !strings.Contains(result, "proxy.example.com") {
				t.Errorf("Expected result to contain proxy domain, but got: %s", result)
			}

			// 验证Markdown格式被保持
			if !strings.Contains(result, "[") || !strings.Contains(result, "]") || !strings.Contains(result, "(") || !strings.Contains(result, ")") {
				t.Errorf("Expected result to maintain markdown format, but got: %s", result)
			}

			// 验证原始链接被替换
			if strings.Contains(result, "google.com") || strings.Contains(result, "example1.com") || strings.Contains(result, "example2.com") || strings.Contains(result, "api.openai.com") {
				t.Errorf("Expected original URLs to be replaced, but got: %s", result)
			}
		})
	}
}

func TestJSONWithMarkdownLinks(t *testing.T) {
	// Setup test configuration
	LinkReplacementEnabled = true
	ReplacementDomain = "proxy.example.com"

	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "Simple JSON with markdown link",
			input:       `{"message": "Check [OpenAI](https://api.openai.com) for details"}`,
			description: "Should replace markdown link in JSON string value",
		},
		{
			name:        "JSON with multiple markdown links",
			input:       `{"content": "Visit [Site 1](https://example1.com) and [Site 2](https://example2.com)"}`,
			description: "Should replace multiple markdown links",
		},
		{
			name:        "JSON with mixed markdown and plain links",
			input:       `{"text": "See [docs](https://docs.example.com) or visit https://example.com directly"}`,
			description: "Should replace both markdown and plain links",
		},
		{
			name:        "Complex JSON with nested markdown links",
			input:       `{"response": {"content": "Please visit [our API](https://api.test.com/v1) for more information"}, "links": ["https://support.test.com"]}`,
			description: "Should handle complex JSON structures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceLinksInJSON(tt.input)

			// 验证结果包含代理域名
			if !strings.Contains(result, "proxy.example.com") {
				t.Errorf("Expected result to contain proxy domain, but got: %s", result)
			}

			// 验证JSON结构保持完整
			if !strings.Contains(result, "{") || !strings.Contains(result, "}") {
				t.Errorf("Expected result to maintain JSON structure, but got: %s", result)
			}

			// 验证Markdown格式在JSON中被保持
			if strings.Contains(tt.input, "[") && strings.Contains(tt.input, "]") {
				if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
					t.Errorf("Expected markdown format to be preserved in JSON, but got: %s", result)
				}
			}

			// 验证原始域名被替换
			originalDomains := []string{"api.openai.com", "example1.com", "example2.com", "docs.example.com", "example.com", "api.test.com", "support.test.com"}
			for _, domain := range originalDomains {
				if strings.Contains(tt.input, domain) {
					// 检查原始URL是否被替换（但允许在base64编码中存在）
					plainDomainPattern := "https://" + domain
					if strings.Contains(result, plainDomainPattern) {
						t.Errorf("Expected original URL %s to be replaced, but it still appears in: %s", plainDomainPattern, result)
					}
				}
			}

			t.Logf("Test: %s", tt.description)
			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
		})
	}
}

func TestEncryptionBasedEncoding(t *testing.T) {
	// Setup test configuration
	LinkReplacementEnabled = true
	ReplacementDomain = "proxy.example.com"

	testCases := []string{
		"https://api.openai.com",
		"https://example.com/api/v1/models",
		"https://docs.example.com/getting-started",
	}

	for _, originalURL := range testCases {
		t.Run(originalURL, func(t *testing.T) {
			// 加密同一个URL多次，应该产生不同的结果（因为使用随机nonce）
			encoded1 := encodeLink(originalURL)
			encoded2 := encodeLink(originalURL)

			parts1 := strings.Split(encoded1, "/redirect/")
			parts2 := strings.Split(encoded2, "/redirect/")

			if len(parts1) != 2 || len(parts2) != 2 {
				t.Fatalf("Invalid encoded URL format")
			}

			encodedPart1 := parts1[1]
			encodedPart2 := parts2[1]

			// 验证每次加密的结果都不同（随机性）
			if encodedPart1 == encodedPart2 {
				t.Errorf("Encryption should produce different results each time due to random nonce")
			}

			// 验证都能正确解密回原始URL
			decoded1, err1 := DecodeLink(encodedPart1)
			decoded2, err2 := DecodeLink(encodedPart2)

			if err1 != nil || err2 != nil {
				t.Fatalf("Failed to decode: %v, %v", err1, err2)
			}

			if decoded1 != originalURL || decoded2 != originalURL {
				t.Errorf("Decoding failed. Expected %s, got %s and %s", originalURL, decoded1, decoded2)
			}

			// 验证编码结果看起来像随机十六进制
			for _, encoded := range []string{encodedPart1, encodedPart2} {
				if len(encoded) < 32 {
					t.Errorf("Encoded string too short: %s", encoded)
				}

				// 验证是有效的十六进制
				for _, char := range encoded {
					if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
						t.Errorf("Invalid hex character: %c in %s", char, encoded)
					}
				}
			}

			t.Logf("Original: %s", originalURL)
			t.Logf("Encoded1: %s", encodedPart1)
			t.Logf("Encoded2: %s", encodedPart2)
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}