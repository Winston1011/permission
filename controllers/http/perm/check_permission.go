package perm

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/perm"
)

func CheckPermission(ctx *gin.Context) {
	var params struct {
		ProductId int64  `json:"productId" form:"productId" binding:"required"`
		AppId     int64  `json:"appId" form:"appId" binding:"required"`
		UserId    int64  `json:"userId" form:"userId" binding:"required"`
		Resource  string `json:"resource" form:"resource" binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorPermissionParamsInvalid)
		return
	}
	checkInput := &perm.CheckInput{
		ProductId: params.ProductId,
		AppId:     params.AppId,
		UserId:    params.UserId,
		Resource:  params.Resource,
	}
	response, err := checkInput.CheckPermission(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
