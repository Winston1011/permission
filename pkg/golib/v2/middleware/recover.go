package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery(handle gin.RecoveryFunc) gin.HandlerFunc {
	// 如果用户没有自定义recovery handler，使用默认的handler
	if handle == nil {
		handle = func(c *gin.Context, err interface{}) {
			// 默认回500, 业务可自定义(handler字需要包括返回的response body即可)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}

	return gin.CustomRecovery(handle)
}
