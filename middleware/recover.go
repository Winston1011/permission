package middleware

import (
	"bytes"
	"io/ioutil"

	"permission/components"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/zlog"
)

func Recover(ctx *gin.Context) {
	defer CatchRecoverRpc(ctx)
	ctx.Next()
}

// 针对rpc接口的处理
func CatchRecoverRpc(c *gin.Context) {
	// panic捕获
	if err := recover(); err != nil {
		// 请求url
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		// 请求报文
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		fields := []zlog.Field{
			zlog.String("logId", zlog.GetLogID(c)),
			zlog.String("requestId", zlog.GetRequestID(c)),
			zlog.String("uri", path),
			zlog.String("refer", c.Request.Referer()),
			zlog.String("clientIp", c.ClientIP()),
			zlog.String("module", env.AppName),
			zlog.String("ua", c.Request.UserAgent()),
			zlog.String("host", c.Request.Host),
		}
		zlog.InfoLogger(c, "Panic[recover]", fields...)

		base.RenderJsonAbort(c, components.ErrorSystemError)
	}
}
