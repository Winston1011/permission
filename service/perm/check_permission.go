package perm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"permission/api"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"permission/pkg/golib/v2/zlog"
)

type CheckInput struct {
	ProductId int64
	AppId     int64
	UserId    int64
	Resource  string
}

type CheckOutput struct {
	Allow bool `json:"allow"`
}

func (ci *CheckInput) CheckPermission(ctx *gin.Context) (CheckOutput, error) {
	err := ci.checkParams()
	if err != nil {
		return CheckOutput{Allow: false}, err
	}
	var userType int8
	// 1. 查看userId 是内网/外网 用户 (同时还要查看userId的有效性)
	infoFromPass, err := api.GetUserInfoByUserId(ctx, ci.AppId, ci.UserId)
	if err != nil {
		zlog.Errorf(ctx, "passport get userinfo failure", err)
		return CheckOutput{Allow: false}, helpers.NewError(components.ErrorApiGetUserInfo, err.Error())
	}
	if infoFromPass.UserId > 0 {
		userType = components.USER_TYPE_OUTER
	}
	userGroup := &m.UserGroup{
		UserId: ci.UserId,
	}
	condition := map[string]interface{}{
		"product_id": ci.ProductId,
		"app_id":     ci.AppId,
		"user_type":  userType,
		"user_id":    ci.UserId,
	}
	userGroupInfo, err := userGroup.GetUserGroupByCondition(ctx, condition)
	if err != nil {
		return CheckOutput{Allow: false}, helpers.NewError(components.ErrorDbSelect, "get userGroup by condition error")
	}
	sub := fmt.Sprintf("%d", userGroupInfo.GroupId)
	dom := fmt.Sprintf("%d:%d", ci.ProductId, ci.AppId)
	obj := ci.Resource
	act := components.CASBIN_ACT_ANY
	e := helpers.Enforcer
	// 判断策略中是否存在
	result, err := e.Enforce(sub, dom, obj, act)
	if err != nil {
		zlog.Errorf(ctx, "casbin check machine does not work err:%s", err)
	}
	return CheckOutput{Allow: result}, err
}

func (ci *CheckInput) checkParams() error {
	if ci.AppId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "appId 不合法")
	}
	if ci.ProductId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "productId 不合法")
	}
	if ci.UserId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "userId 不合法")
	}
	if len(ci.Resource) < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "resource 不合法")
	}
	return nil
}
