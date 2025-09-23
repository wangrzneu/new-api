package main

import (
	"fmt"
	"net/http"
	"one-api/types"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 示例路由：演示中英文错误消息
	r.GET("/demo-error", func(c *gin.Context) {
		// 模拟一个错误场景
		err := types.NewI18nError(
			c,
			fmt.Errorf("model not available"),
			types.ErrorCodeModelNotFound,
		)

		// 获取国际化错误消息
		i18nMsg := err.ToI18nMessage(c)

		// 返回国际化错误响应
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    i18nMsg.Code,
				"message": i18nMsg.Message,
				"lang":    i18nMsg.Lang,
			},
		})
	})

	// 示例路由：演示不同错误类型的国际化
	r.GET("/demo-quota-error", func(c *gin.Context) {
		err := types.NewI18nError(
			c,
			fmt.Errorf("quota exceeded"),
			types.ErrorCodeInsufficientUserQuota,
		)

		i18nMsg := err.ToI18nMessage(c)

		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    i18nMsg.Code,
				"message": i18nMsg.Message,
				"lang":    i18nMsg.Lang,
			},
		})
	})

	// 示例路由：演示 OpenAI 格式的国际化错误
	r.GET("/demo-openai-error", func(c *gin.Context) {
		err := types.NewI18nOpenAIError(
			c,
			fmt.Errorf("invalid request format"),
			types.ErrorCodeInvalidRequest,
			http.StatusBadRequest,
		)

		// 转换为 OpenAI 错误格式
		openaiError := err.ToOpenAIError()

		c.JSON(http.StatusBadRequest, gin.H{
			"error": openaiError,
		})
	})

	fmt.Println("国际化错误消息演示服务器启动在 :8080")
	fmt.Println("测试端点:")
	fmt.Println("  英文: curl -H 'Accept-Language: en' http://localhost:8080/demo-error")
	fmt.Println("  中文: curl -H 'Accept-Language: zh-CN' http://localhost:8080/demo-error")
	fmt.Println("  查询参数: curl http://localhost:8080/demo-error?lang=zh")

	r.Run(":8080")
}

// 演示直接使用国际化函数
func demonstrateI18nFunctions() {
	// 创建一个模拟的 Gin Context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// 模拟中文请求
	c.Request = &http.Request{
		Header: http.Header{
			"Accept-Language": []string{"zh-CN,zh;q=0.9,en;q=0.8"},
		},
	}

	// 获取错误消息
	errMsg := types.GetErrorMessage(c, types.ErrorCodeModelNotFound)
	fmt.Printf("中文错误消息: %+v\n", errMsg)

	// 模拟英文请求
	c.Request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	errMsg = types.GetErrorMessage(c, types.ErrorCodeModelNotFound)
	fmt.Printf("英文错误消息: %+v\n", errMsg)
}