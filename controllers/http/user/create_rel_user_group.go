package user

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/user"
)

func CreateRelUserGroup(ctx *gin.Context) {
	var params struct {
		ProductId  int64 `json:"productId" form:"productId" binding:"required"`
		AppId      int64 `json:"appId" form:"appId" binding:"required"`
		UserType   int8  `json:"userType" form:"userType" binding:"required"`
		UserId     int64 `json:"userId" form:"userId" binding:"required"` // 通过uid来添加
		GroupId    int64 `json:"groupId" form:"groupId" binding:"required"`
		Status     int8  `json:"status" form:"status"`
		OperateUid int64 `json:"operateUid" form:"operateUid" binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorUserGroupParamsInvalid)
		return
	}
	userGroupInput := &user.RCreateInput{
		ProductId:  params.ProductId,
		AppId:      params.AppId,
		UserType:   params.UserType,
		UserId:     params.UserId,
		GroupId:    params.GroupId,
		Status:     params.Status,
		OperateUid: params.OperateUid,
	}
	response, err := userGroupInput.CreateUserGroup(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
