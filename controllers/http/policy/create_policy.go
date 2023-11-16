package policy

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
	"permission/service/policy"
)

func CreatePolicy(ctx *gin.Context) {
	var params struct {
		GroupId        int64  `json:"groupId" form:"groupId" binding:"required"`
		AppId          int64  `json:"appId" form:"appId" binding:"required"`
		ProductId      int64  `json:"productId" form:"productId" binding:"required"`
		Resource       string `json:"resource" form:"resource" binding:"required"`
		PermissionType string `json:"permissionType" form:"permissionType" binding:"required"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		zlog.Warnf(ctx, "json params reflection failure err:%v", err)
		base.RenderJsonFail(ctx, components.ErrorPolicyParamsInvalid)
		return
	}
	policyInput := &policy.PCreateInput{
		GroupId:        params.GroupId,
		AppId:          params.AppId,
		ProductId:      params.ProductId,
		Resource:       params.Resource,
		PermissionType: params.PermissionType,
	}
	response, err := policyInput.CreatePolicy(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
