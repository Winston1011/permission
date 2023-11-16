package middleware

import (
	"github.com/gin-gonic/gin"
	"permission/pkg/antispam"
	"permission/pkg/golib/v2/base"
)

func AppCheck(ctx *gin.Context) {
	if err := antispam.AppCheck(ctx); err != nil {
		base.RenderJsonAbort(ctx, err)
		return
	}
	ctx.Next()
}

func SdkCheck(ctx *gin.Context) {
	if err := antispam.SdkCheck(ctx); err != nil {
		base.RenderJsonAbort(ctx, err)
		return
	}
	ctx.Next()
}
