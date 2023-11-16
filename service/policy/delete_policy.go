package policy

import (
	"git.zuoyebang.cc/pkg/golib/v2/zlog"
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
)

func DeletePolicyById(ctx *gin.Context, id int64) (bool, error) {
	if id < 0 {
		return false, helpers.NewError(components.ErrorPolicyParamsInvalid, "id 不合法")
	}
	casbinRule := &m.CasbinRule{
		ID: id,
	}
	_, err := casbinRule.DeleteCasbinRule(ctx)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbDelete, "delete casbinRule by id failure")
	}
	err = helpers.Enforcer.LoadPolicy()
	if err != nil {
		zlog.Warnf(ctx, "casbin reload policy failure", err)
	}
	return true, nil

}
