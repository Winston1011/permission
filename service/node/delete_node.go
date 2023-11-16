package node

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
)

func DeleteNode(ctx *gin.Context, id int64) (bool, error) {
	if id < 0 {
		return false, helpers.NewError(components.ErrorNodeParamsInvalid, "id 不合法")
	}
	node := &m.Node{ID: id}
	_, err := node.DeleteNodeById(ctx)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbDelete, "delete node by id failure")
	}
	return true, nil
}
