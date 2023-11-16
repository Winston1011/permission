package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/group"
)

func DeleteGroup(ctx *gin.Context) {
	var params struct {
		UserId  int64 `json:"userId" form:"userId" binding:"required"`
		GroupId int64 `json:"groupId" form:"groupId" binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "group delete params invalid err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorGroupParamsInvalid)
		return
	}
	groupInput := &group.GDeleteInput{
		UserId:  params.UserId,
		GroupId: params.GroupId,
	}
	response, err := groupInput.DeleteGroup(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
