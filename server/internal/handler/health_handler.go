package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleHealth handles health check requests
// HandleHealth 处理健康检查请求
func HandleHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
