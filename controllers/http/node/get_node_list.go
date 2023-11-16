package node

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	n "permission/service/node"
)

func GetNodeList(ctx *gin.Context) {
	var params struct {
		AppId     int64 `json:"appId" form:"appId" binding:"required"`
		ProductId int64 `json:"productId" form:"productId" binding:"required"`
		GroupId   int64 `json:"groupId" form:"groupId"`
		NodeType  int8  `json:"nodeType" form:"nodeType"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorNodeParamsInvalid)
		return
	}
	nodeInput := &n.NodeListInput{
		AppId:     params.AppId,
		ProductId: params.ProductId,
		GroupId:   params.GroupId,
		NodeType:  params.NodeType,
	}
	response, err := nodeInput.GetNodeList(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
