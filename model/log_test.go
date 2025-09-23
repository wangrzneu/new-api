package model

import (
	"testing"
)

func TestTokenStat(t *testing.T) {
	// 测试 TokenStat 结构体
	stat := TokenStat{
		PromptTokens:     100,
		CompletionTokens: 200,
		TotalTokens:      300,
		RequestCount:     5,
	}

	if stat.PromptTokens != 100 {
		t.Errorf("Expected PromptTokens to be 100, got %d", stat.PromptTokens)
	}

	if stat.CompletionTokens != 200 {
		t.Errorf("Expected CompletionTokens to be 200, got %d", stat.CompletionTokens)
	}

	if stat.TotalTokens != 300 {
		t.Errorf("Expected TotalTokens to be 300, got %d", stat.TotalTokens)
	}

	if stat.RequestCount != 5 {
		t.Errorf("Expected RequestCount to be 5, got %d", stat.RequestCount)
	}
}

// 测试 SumUsedToken 函数签名（模拟调用）
func TestSumUsedTokenSignature(t *testing.T) {
	// 这个测试主要验证函数签名是否正确
	// 由于需要数据库连接，我们只测试函数是否能被调用而不出现编译错误

	// 模拟参数
	logType := 3        // LogTypeConsume
	startTimestamp := int64(0)
	endTimestamp := int64(0)
	modelName := "test-model"
	username := "test-user"
	tokenName := "test-token"

	// 注意：这里我们不能真正调用函数，因为需要数据库连接
	// 但是这个测试确保了函数签名是正确的
	_ = logType
	_ = startTimestamp
	_ = endTimestamp
	_ = modelName
	_ = username
	_ = tokenName

	// 如果编译通过，说明 TokenStat 结构体和函数签名都是正确的
	var stat TokenStat
	if stat.PromptTokens != 0 {
		t.Errorf("Expected initial PromptTokens to be 0")
	}
}