package node

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"time"
)

type NCreateInput struct {
	AppId     int64
	ProductId int64
	Label     string
	Resource  string
	ParentId  int64
	UserId    int64
	IsShow    int8
	NodeType  int8
}

func (nc *NCreateInput) CreateNode(ctx *gin.Context) (bool, error) {
	if err := nc.checkParams(); err != nil {
		return false, err
	}
	node := &m.Node{
		ProductID:  nc.ProductId,
		AppID:      nc.AppId,
		Label:      nc.Label,
		Resource:   nc.Resource,
		IsShow:     nc.IsShow,
		ParentID:   nc.ParentId,
		NodeType:   nc.NodeType,
		CreateUid:  nc.UserId,
		UpdateUid:  nc.UserId,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}
	condition := map[string]interface{}{
		"product_id": nc.ProductId,
		"app_id":     nc.AppId,
		"label":      nc.Label,
		"node_type":  nc.NodeType,
		"resource":   nc.Resource,
		"parent_id":  nc.ParentId,
	}
	nodeInfo, _ := node.GetNodeByCondition(ctx, condition)
	if nodeInfo.ID > 0 {
		return false, helpers.NewError(components.ErrorDbInsert, "节点资源已存在")
	}
	err := node.InsertNode(ctx)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbInsert, "insert node failure")
	}
	return true, nil
}

func (nc *NCreateInput) checkParams() error {
	if nc.AppId < 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "appId 不合法")
	}
	if nc.ProductId < 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "productId 不合法")
	}
	if nc.UserId < 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "userId 不合法")
	}
	if len(nc.Label) <= 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "label 不合法")
	}
	if len(nc.Resource) <= 0 {
		return helpers.NewError(components.ErrorNodeParamsInvalid, "resource 不合法")
	}
	return nil
}
