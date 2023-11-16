package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/group"
)

func Updategroup(ctx *gin.Context) {
	var params struct {
		ProductId   int64   `json:"productId" form:"productId" binding:"required"`
		AppId       int64   `json:"appId" form:"appId" binding:"required"`
		GroupId     int64   `json:"groupId" form:"groupId" binding:"required"`
		UserId      int64   `json:"userId" form:"userId" binding:"required"`
		GroupName   string  `json:"groupName" form:"groupName" binding:"required"`
		MenuList    []int64 `json:"menuList" form:"menuList" binding:"required"`
		NodeList    []int64 `json:"nodeList" form:"nodeList" binding:"required"`
		GroupStatus int8    `json:"groupStatus" form:"groupStatus"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorGroupParamsInvalid)
		return
	}
	groupInput := &group.GUpdateInput{
		ProductId:   params.ProductId,
		AppId:       params.AppId,
		GroupId:     params.GroupId,
		UserId:      params.UserId,
		GroupName:   params.GroupName,
		GroupStatus: params.GroupStatus,
		NodeList:    params.NodeList,
		MenuList:    params.MenuList,
	}
	response, err := groupInput.UpdateGroup(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
