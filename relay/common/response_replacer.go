package common

import (
	"bytes"
	"io"
	"net/http"
	"one-api/common"
	"strconv"
	"strings"
)

// ApplyLinkReplacement 应用链接替换功能的统一入口函数
func ApplyLinkReplacement(content string) string {
	if !common.LinkReplacementEnabled || common.ReplacementDomain == "" {
		return content
	}

	// 检查是否包含JSON数据
	if strings.Contains(content, "{") && strings.Contains(content, "}") {
		return common.ReplaceLinksInJSON(content)
	} else {
		return common.ReplaceLinksInText(content)
	}
}

// ApplyLinkReplacementToBytes 处理字节数据的链接替换
func ApplyLinkReplacementToBytes(data []byte) []byte {
	if !common.LinkReplacementEnabled || common.ReplacementDomain == "" {
		return data
	}

	processedContent := ApplyLinkReplacement(string(data))
	return []byte(processedContent)
}

// ProcessResponseWithLinkReplacement 处理HTTP响应并替换其中的链接
func ProcessResponseWithLinkReplacement(resp *http.Response) error {
	if !common.LinkReplacementEnabled || common.ReplacementDomain == "" {
		return nil
	}

	// 只处理文本类型的响应
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "text/") {
		return nil
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// 应用链接替换
	processedBody := ApplyLinkReplacement(string(body))

	// 重新设置响应体
	resp.Body = io.NopCloser(bytes.NewReader([]byte(processedBody)))
	resp.ContentLength = int64(len(processedBody))
	resp.Header.Set("Content-Length", strconv.Itoa(len(processedBody)))

	return nil
}

// ProcessStreamResponseWithLinkReplacement 处理流式响应并替换其中的链接
func ProcessStreamResponseWithLinkReplacement(data []byte) []byte {
	return ApplyLinkReplacementToBytes(data)
}