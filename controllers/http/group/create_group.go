package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/group"
)

func CreateGroup(ctx *gin.Context) {
	var params struct {
		ProductId int64  `json:"productId" form:"productId" binding:"required"`
		AppId     int64  `json:"appId" form:"appId" binding:"required"`
		UserId    int64  `json:"userId" form:"userId" binding:"required"`
		GroupName string `json:"groupName" form:"groupName" binding:"required"`
		ParentId  int64  `json:"parentId" form:"parentId"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorGroupParamsInvalid)
		return
	}
	groupInput := &group.GCreateInput{
		UserId:    params.UserId,
		ProductId: params.ProductId,
		AppId:     params.AppId,
		GroupName: params.GroupName,
		ParentId:  params.ParentId,
	}
	response, err := groupInput.CreateGroup(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
