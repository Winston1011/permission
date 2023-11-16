package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"time"
)

type GDeleteInput struct {
	UserId  int64
	GroupId int64
}

func (gd *GDeleteInput) DeleteGroup(ctx *gin.Context) (bool, error) {
	if err := gd.checkParams(); err != nil {
		return false, err
	}
	group := &m.Group{}
	updatedFields := map[string]interface{}{
		"status":      components.GROUP_STATUS_DELETED,
		"update_uid":  gd.UserId,
		"update_time": time.Now().Unix(),
	}
	_, err := group.UpdateGroupById(ctx, gd.GroupId, updatedFields, nil)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbUpdate, "delete group by id failure")
	}
	return true, nil
}

func (gd *GDeleteInput) checkParams() error {
	if gd.UserId <= 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "userId 不合法")
	}
	if gd.GroupId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "groupId 不合法")
	}
	return nil
}
