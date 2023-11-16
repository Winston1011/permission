package user

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"time"
)

type RCreateInput struct {
	ProductId  int64
	AppId      int64
	UserType   int8
	UserId     int64
	GroupId    int64
	Status     int8
	OperateUid int64
}

func (rc *RCreateInput) CreateUserGroup(ctx *gin.Context) (bool, error) {
	if err := rc.checkParams(); err != nil {
		return false, err
	}
	userGroup := &m.UserGroup{
		ProductId:  rc.ProductId,
		AppId:      rc.AppId,
		UserType:   rc.UserType,
		UserId:     rc.UserId,
		GroupId:    rc.GroupId,
		CreateUid:  rc.OperateUid,
		UpdateUid:  rc.OperateUid,
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
		Status:     rc.Status,
	}
	condition := map[string]interface{}{
		"product_id": rc.ProductId,
		"app_id":     rc.AppId,
		"user_type":  rc.UserType,
		"user_id":    rc.UserId,
	}
	userGroupInfo, err1 := userGroup.GetUserGroupByCondition(ctx, condition)
	if err1 != nil {
		return false, helpers.NewError(components.ErrorDbSelect, "get all userGroupList by condition failure")
	}
	userGroup.ID = userGroupInfo.ID
	_, err2 := userGroup.UpsertUserGroup(ctx)
	if err2 != nil {
		return false, helpers.NewError(components.ErrorDbUpdate, "upsert userGroup failure")
	}
	return true, nil
}

func (rc *RCreateInput) checkParams() error {
	if rc.ProductId < 0 {
		return helpers.NewError(components.ErrorUserGroupParamsInvalid, "productId 不合法")
	}
	if rc.AppId < 0 {
		return helpers.NewError(components.ErrorUserGroupParamsInvalid, "appId 不合法")
	}
	if rc.UserType < 0 {
		return helpers.NewError(components.ErrorUserGroupParamsInvalid, "userType 不合法")
	}
	if rc.GroupId < 0 {
		return helpers.NewError(components.ErrorUserGroupParamsInvalid, "groupId 不合法")
	}
	if rc.Status < 0 {
		return helpers.NewError(components.ErrorUserGroupParamsInvalid, "status 不合法")
	}
	if rc.OperateUid < 0 {
		return helpers.NewError(components.ErrorUserGroupParamsInvalid, "operatedUid 不合法")
	}
	return nil
}
