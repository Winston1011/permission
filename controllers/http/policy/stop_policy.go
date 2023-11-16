package policy

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/policy"
)

func StopPolicy(ctx *gin.Context) {
	var params struct {
		Id int64 `json:"id" form:"id" binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorPolicyParamsInvalid)
		return
	}
	response, err := policy.StopPolicyById(ctx, params.Id)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
