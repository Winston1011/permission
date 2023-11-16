package policy

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
)

type ListOutput struct {
	PolicyList []m.CasbinRule `json:"policyList"`
}

type ListInput struct {
}

func (li *ListInput) GetPolicyList(ctx *gin.Context) (ListOutput, error) {
	policy := &m.CasbinRule{}
	policyList, err := policy.GetCasbinRulesListByConds(ctx, nil)
	response := ListOutput{
		PolicyList: policyList,
	}
	if err != nil {
		return response, helpers.NewError(components.ErrorDbSelect, "get all policyList failure")
	}
	return response, nil
}
