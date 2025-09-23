package router

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"one-api/common"
	"one-api/controller"

	"github.com/gin-gonic/gin"
)

func SetRouter(router *gin.Engine, buildFS embed.FS, indexPage []byte) {
	SetApiRouter(router)
	SetDashboardRouter(router)
	SetRelayRouter(router)
	SetVideoRouter(router)
	SetAsyncImageRouter(router)

	// 添加重定向路由处理编码的链接
	router.GET("/redirect/:encoded", controller.RedirectHandler)
	frontendBaseUrl := os.Getenv("FRONTEND_BASE_URL")
	if common.IsMasterNode && frontendBaseUrl != "" {
		frontendBaseUrl = ""
		common.SysLog("FRONTEND_BASE_URL is ignored on master node")
	}
	if frontendBaseUrl == "" {
		SetWebRouter(router, buildFS, indexPage)
	} else {
		frontendBaseUrl = strings.TrimSuffix(frontendBaseUrl, "/")
		router.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", frontendBaseUrl, c.Request.RequestURI))
		})
	}
}
