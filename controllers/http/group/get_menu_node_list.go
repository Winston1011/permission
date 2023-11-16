package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/group"
)

func GetMenuNodeList(ctx *gin.Context) {
	var params struct {
		AppId     int64 `json:"appId" form:"appId" binding:"required"`
		ProductId int64 `json:"productId" form:"productId" binding:"required"`
		UserId    int64 `json:"userId" form:"userId"  binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorGroupParamsInvalid)
		return
	}
	getInput := &group.GetListInput{
		AppId:     params.AppId,
		ProductId: params.ProductId,
		UserId:    params.UserId,
	}
	response, err := getInput.GetMenuNodeList(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
