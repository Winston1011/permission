package node

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"time"
)

type NUpdateInput struct {
	Id       int64
	Label    string
	Resource string
	ParentId int64
	UserId   int64
}

func (nu *NUpdateInput) UpdateNode(ctx *gin.Context) (bool, error) {
	if err := nu.checkParams(); err != nil {
		return false, err
	}
	node := &m.Node{
		ID: nu.Id,
	}
	updatedFields := map[string]interface{}{
		"label":       nu.Label,
		"parent_id":   nu.ParentId,
		"resource":    nu.Resource,
		"update_uid":  nu.UserId,
		"update_time": time.Now().Unix(),
	}
	_, err := node.UpdateNodeById(ctx, nu.Id, updatedFields)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbInsert, "update node by id failure")
	}
	return true, nil
}

func (nu *NUpdateInput) checkParams() error {
	if nu.Id < 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "id 不合法")
	}
	if len(nu.Label) < 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "label 不合法")
	}
	if len(nu.Resource) <= 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "resource 不合法")
	}
	if nu.UserId <= 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "userId 不合法")
	}
	return nil
}
