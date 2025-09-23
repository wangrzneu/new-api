package controller

import (
	"net/http"
	"one-api/common"

	"github.com/gin-gonic/gin"
)

func RedirectHandler(c *gin.Context) {
	encodedURL := c.Param("encoded")
	if encodedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing encoded URL"})
		return
	}

	// 解码原始URL
	originalURL, err := common.DecodeLink(encodedURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid encoded URL"})
		return
	}

	// 重定向到原始URL
	c.Redirect(http.StatusFound, originalURL)
}