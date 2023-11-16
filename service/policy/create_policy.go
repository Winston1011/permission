package policy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
)

type PCreateInput struct {
	GroupId        int64
	AppId          int64
	ProductId      int64
	Resource       string
	PermissionType string
}

func (pi *PCreateInput) CreatePolicy(ctx *gin.Context) (bool, error) {
	if err := pi.checkParams(); err != nil {
		return false, err
	}
	policy := &m.CasbinRule{
		Ptype:           components.CASBIN_RULE_PTYPE,
		GroupId:         fmt.Sprintf("%d", pi.GroupId),
		ProductAppField: fmt.Sprintf("%d:%d", pi.ProductId, pi.AppId),
		Resource:        pi.Resource,
		PermissionType:  pi.PermissionType,
		Status:          components.POLICY_STATUS_ALLOW,
	}
	condition := map[string]interface{}{
		"ptype": components.CASBIN_RULE_PTYPE,
		"v0":    fmt.Sprintf("%d", pi.GroupId),
		"v1":    fmt.Sprintf("%d:%d", pi.ProductId, pi.AppId),
		"v2":    pi.Resource,
		"v3":    pi.PermissionType,
	}
	policyInfo, _ := policy.GetCasbinRulesByConds(ctx, condition)
	if policyInfo.ID > 0 {
		return false, helpers.NewError(components.ErrorDbInsert, "校验规则已存在")
	}
	_, err := helpers.Enforcer.AddPolicy(fmt.Sprintf("%d", pi.GroupId), fmt.Sprintf("%d:%d", pi.ProductId, pi.AppId), pi.Resource, pi.PermissionType, components.POLICY_STATUS_ALLOW)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbInsert, "insert policy failure")
	}
	return true, nil
}

func (pi *PCreateInput) checkParams() error {
	if pi.GroupId < 0 {
		return helpers.NewError(components.ErrorPolicyParamsInvalid, "groupId 不合法")
	}
	if pi.AppId < 0 {
		return helpers.NewError(components.ErrorPolicyParamsInvalid, "appId 不合法")
	}
	if pi.ProductId < 0 {
		return helpers.NewError(components.ErrorPolicyParamsInvalid, "productId 不合法")
	}
	return nil
}
