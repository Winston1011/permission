package node

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/node"
)

func CreateNode(ctx *gin.Context) {
	var params struct {
		AppId     int64  `json:"appId" form:"appId" binding:"required"`
		ProductId int64  `json:"productId" form:"productId" binding:"required"`
		Label     string `json:"label" form:"label" binding:"required"`
		Resource  string `json:"resource" form:"resource" binding:"required"`
		IsShow    int8   `json:"isShow" form:"isShow"`
		NodeType  int8   `json:"nodeType" form:"nodeType"`
		UserId    int64  `json:"userId" form:"userId" binding:"required"`
		ParentId  int64  `json:"parentId" form:"parentId"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorNodeParamsInvalid)
		return
	}
	nodeInput := &node.NCreateInput{
		AppId:     params.AppId,
		ProductId: params.ProductId,
		Label:     params.Label,
		Resource:  params.Resource,
		IsShow:    params.IsShow,
		ParentId:  params.ParentId,
		UserId:    params.UserId,
		NodeType:  params.NodeType,
	}
	response, err := nodeInput.CreateNode(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}

}
