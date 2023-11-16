package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"time"
)

type GCreateInput struct {
	UserId    int64
	ProductId int64
	AppId     int64
	GroupName string
	ParentId  int64
}

func (gi *GCreateInput) CreateGroup(ctx *gin.Context) (bool, error) {
	if err := gi.checkParams(); err != nil {
		return false, err
	}
	group := &m.Group{
		ProductID:  gi.ProductId,
		AppID:      gi.AppId,
		GroupName:  gi.GroupName,
		CreateUid:  gi.UserId,
		UpdateUid:  gi.UserId,
		ParentId:   gi.ParentId,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
	}
	condition := map[string]interface{}{
		"product_id": gi.ProductId,
		"app_id":     gi.AppId,
		"group_name": gi.GroupName,
		"status":     components.GROUP_STATUS_ACTIVE,
	}
	groupInfo, _ := group.GetGroupByConds(ctx, condition)
	if groupInfo.ID > 0 {
		return false, helpers.NewError(components.ErrorDbInsert, "权限组已存在")
	}
	group.ID = groupInfo.ID
	err := group.InsertGroup(ctx)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbInsert, "insert group failure")
	}
	return true, nil
}

func (gi *GCreateInput) checkParams() error {
	if gi.UserId <= 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "userId 不合法")
	}
	if gi.ProductId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "productId 不合法 ")
	}
	if gi.AppId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "appId 不合法")
	}
	return nil
}
