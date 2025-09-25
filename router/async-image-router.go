package router

import (
	"one-api/controller"
	"one-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetAsyncImageRouter(router *gin.Engine) {
	imageV1Router := router.Group("/v1")
	imageV1Router.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		imageV1Router.POST("/images/generations/async", RelayAsyncImageTask)
		imageV1Router.GET("/images/generations/async/:task_id", RelayAsyncImageTask)
	}
}

func RelayAsyncImageTask(c *gin.Context) {
	c.Set("is_async_image", true)
	controller.RelayTask(c)
}
