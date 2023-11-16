package node

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/node"
)

func UpdateNode(ctx *gin.Context) {
	var params struct {
		Id       int64  `json:"id" form:"id" binding:"required"`
		Label    string `json:"label" form:"label" binding:"required"`
		Resource string `json:"resource" form:"resource" binding:"required"`
		ParentId int64  `json:"parentId" form:"parentId"`
		UserId   int64  `json:"userId" form:"userId" binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorNodeParamsInvalid)
		return
	}
	nodeInput := &node.NUpdateInput{
		Id:       params.Id,
		Label:    params.Label,
		Resource: params.Resource,
		ParentId: params.ParentId,
		UserId:   params.UserId,
	}
	response, err := nodeInput.UpdateNode(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
