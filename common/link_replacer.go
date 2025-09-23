package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// replaceLinksInText 替换文本中的链接
func ReplaceLinksInText(text string) string {
	if !LinkReplacementEnabled || ReplacementDomain == "" {
		return text
	}

	// 首先处理Markdown格式的链接 [text](url)
	markdownLinkRegex := regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\)]+)\)`)
	text = markdownLinkRegex.ReplaceAllStringFunc(text, func(match string) string {
		// 提取链接文本和URL
		submatch := markdownLinkRegex.FindStringSubmatch(match)
		if len(submatch) == 3 {
			linkText := submatch[1]
			originalURL := submatch[2]
			encodedURL := encodeLink(originalURL)
			return fmt.Sprintf("[%s](%s)", linkText, encodedURL)
		}
		return match
	})

	// 然后处理普通的HTTP和HTTPS链接
	linkRegex := regexp.MustCompile(`https?://[^\s\)\]]+`)
	return linkRegex.ReplaceAllStringFunc(text, func(link string) string {
		return encodeLink(link)
	})
}

// ReplaceLinksInJSON 替换JSON中的链接
func ReplaceLinksInJSON(jsonStr string) string {
	if !LinkReplacementEnabled || ReplacementDomain == "" {
		return jsonStr
	}

	// 匹配JSON字符串值中的链接
	// 这个正则匹配包含HTTP/HTTPS链接的完整JSON字符串值
	linkRegex := regexp.MustCompile(`"([^"]*)"`)

	return linkRegex.ReplaceAllStringFunc(jsonStr, func(match string) string {
		// 提取引号内的内容
		content := match[1 : len(match)-1] // 去掉前后的引号

		// 如果内容包含链接，则进行替换
		if strings.Contains(content, "http://") || strings.Contains(content, "https://") {
			replacedContent := ReplaceLinksInText(content)
			return `"` + replacedContent + `"`
		}

		// 如果不包含链接，返回原内容
		return match
	})
}

// getEncryptionKey 生成加密密钥
func getEncryptionKey() []byte {
	// 使用系统常量和替换域名生成密钥
	keySource := fmt.Sprintf("%s-%s-%s", CryptoSecret, ReplacementDomain, "link-encryption")
	hash := sha256.Sum256([]byte(keySource))
	return hash[:]
}

// encodeLink 编码单个链接
func encodeLink(originalLink string) string {
	// 如果链接已经是我们的代理链接，则不再处理
	if strings.Contains(originalLink, ReplacementDomain+"/redirect/") {
		return originalLink
	}

	// 解析原始URL以确保格式正确
	_, err := url.Parse(originalLink)
	if err != nil {
		// 如果解析失败，返回原链接
		return originalLink
	}

	// 使用AES加密链接
	encryptedData, err := encryptURL(originalLink)
	if err != nil {
		// 如果加密失败，回退到原链接
		return originalLink
	}

	// 构建新的链接
	newLink := fmt.Sprintf("https://%s/redirect/%s",
		strings.TrimSuffix(ReplacementDomain, "/"),
		encryptedData)

	return newLink
}

// encryptURL 使用AES加密URL
func encryptURL(originalURL string) (string, error) {
	key := getEncryptionKey()

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(originalURL), nil)

	// 转换为hex字符串，看起来更像随机ID
	return hex.EncodeToString(ciphertext), nil
}

// DecodeLink 解码链接（用于处理重定向）
func DecodeLink(encodedPath string) (string, error) {
	// 先尝试新的加密方式解码
	if originalURL, err := decryptURL(encodedPath); err == nil {
		return originalURL, nil
	}

	// 如果新方式失败，尝试旧的base64方式（向后兼容）
	decodedBytes, err := base64.URLEncoding.DecodeString(encodedPath)
	if err != nil {
		return "", fmt.Errorf("failed to decode URL")
	}
	return string(decodedBytes), nil
}

// decryptURL 使用AES解密URL
func decryptURL(encryptedHex string) (string, error) {
	key := getEncryptionKey()

	// 从hex字符串转换回字节
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 检查密文长度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// 提取nonce和实际密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}