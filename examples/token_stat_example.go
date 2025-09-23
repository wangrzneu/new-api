package main

import (
	"encoding/json"
	"fmt"
	"one-api/model"
)

func main() {
	// 演示新的 TokenStat 结构体
	tokenStat := model.TokenStat{
		PromptTokens:     1500,
		CompletionTokens: 800,
		TotalTokens:      2300,
		RequestCount:     25,
	}

	fmt.Println("=== Token 统计信息示例 ===")
	fmt.Printf("Prompt Tokens: %d\n", tokenStat.PromptTokens)
	fmt.Printf("Completion Tokens: %d\n", tokenStat.CompletionTokens)
	fmt.Printf("Total Tokens: %d\n", tokenStat.TotalTokens)
	fmt.Printf("Request Count: %d\n", tokenStat.RequestCount)

	// 演示 API 响应格式
	apiResponse := map[string]interface{}{
		"success": true,
		"message": "",
		"data": map[string]interface{}{
			// 原有的统计数据
			"quota": 10000,
			"rpm":   50,
			"tpm":   2500,
			// 新增的 token 统计数据
			"prompt_tokens":     tokenStat.PromptTokens,
			"completion_tokens": tokenStat.CompletionTokens,
			"total_tokens":      tokenStat.TotalTokens,
			"request_count":     tokenStat.RequestCount,
		},
	}

	fmt.Println("\n=== API 响应格式示例 ===")
	jsonData, _ := json.MarshalIndent(apiResponse, "", "  ")
	fmt.Println(string(jsonData))

	fmt.Println("\n=== 使用场景说明 ===")
	fmt.Println("1. prompt_tokens: 用户输入的 token 数量")
	fmt.Println("2. completion_tokens: AI 生成的 token 数量")
	fmt.Println("3. total_tokens: 总 token 数量 (prompt + completion)")
	fmt.Println("4. request_count: 总请求次数")
	fmt.Println("5. quota: 消费的配额")
	fmt.Println("6. rpm: 每分钟请求数")
	fmt.Println("7. tpm: 每分钟 token 数")

	fmt.Println("\n=== API 端点 ===")
	fmt.Println("GET /api/log/stat - 获取管理员统计")
	fmt.Println("GET /api/log/self/stat - 获取用户自己的统计")
	fmt.Println("\n支持的查询参数:")
	fmt.Println("- type: 日志类型")
	fmt.Println("- start_timestamp: 开始时间戳")
	fmt.Println("- end_timestamp: 结束时间戳")
	fmt.Println("- model_name: 模型名称")
	fmt.Println("- username: 用户名")
	fmt.Println("- token_name: Token 名称")
}