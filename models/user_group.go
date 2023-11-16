package models

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"permission/components"
	"permission/helpers"
	"time"
)

type UserGroup struct {
	ID         int64 `json:"id" gorm:"primary_key;column:id"`
	ProductId  int64 `json:"productId" gorm:"column:product_id"`
	AppId      int64 `json:"appId" gorm:"column:app_id"`
	UserType   int8  `json:"userType" gorm:"column:user_type"`
	UserId     int64 `json:"userId" gorm:"column:user_id" `
	GroupId    int64 `json:"groupId" gorm:"column:group_id" `
	Status     int8  `json:"status" gorm:"column:status"`
	CreateUid  int64 `json:"createUid" gorm:"column:create_uid" `
	UpdateUid  int64 `json:"updateUid" gorm:"column:update_uid" `
	CreateTime int64 `json:"createTime" gorm:"column:create_time" `
	UpdateTime int64 `json:"updateTime"gorm:"column:update_time" `
}

func (ug *UserGroup) TableName() string {
	fmt.Println(ug.UserId)
	return components.TABLE_PREX + fmt.Sprintf("%s%d", "rel_user_group", ug.UserId%16)
}

func (ug *UserGroup) InsertUserGroup(ctx *gin.Context) (err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Table(ug.TableName()).Create(ug).Error
	if err != nil {
		return components.ErrorDbInsert.Wrap(err)
	}
	return nil
}

func (ug *UserGroup) UpsertUserGroup(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Table(ug.TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"app_id", "group_id", "user_type", "status", "update_uid", "update_time"}),
	}).Create(ug)
	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbUpsert.Wrap(err)
		return rows, err
	}
	return rows, nil
}

func (ug *UserGroup) UpdateUserGroupById(ctx *gin.Context, fields map[string]interface{}) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	fields["update_time"] = time.Now().Unix()
	result := db.WithContext(ctx).Table(ug.TableName()).Model(ug).Updates(fields)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbUpdate.Wrap(err)
	}
	return rows, nil
}

func (ug *UserGroup) GetUserGroupById(ctx *gin.Context, id int64) (userGroup UserGroup, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Table(ug.TableName()).Where("`id` = ?", id).Take(&userGroup).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return userGroup, components.ErrorDbSelect.Wrap(err)
	}
	return userGroup, nil
}

func (ug *UserGroup) GetUserGroupByCondition(ctx *gin.Context, condition map[string]interface{}) (userGroup UserGroup, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Table(ug.TableName()).Where(condition).Take(&userGroup).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return userGroup, components.ErrorDbSelect.Wrap(err)
	}
	return userGroup, nil
}

func (ug *UserGroup) GetUserGroupListByPage(ctx *gin.Context, option *Option, page *NormalPage) (userGroups []UserGroup, cnt int, err error) {
	if !option.IsNeedCnt && !option.IsNeedList {
		return userGroups, cnt, nil
	}
	db := helpers.MysqlClientPermission.WithContext(ctx).Table(ug.TableName()).Model(&UserGroup{})
	if option.IsNeedCnt {
		var c int64
		db = db.Count(&c)
		cnt = int(c)
	}
	if option.IsNeedList {
		db = db.Scopes(NormalPaginate(page)).Find(&userGroups)
	}
	if db.Error != nil {
		return userGroups, cnt, components.ErrorDbSelect.Wrap(db.Error)
	}
	return userGroups, cnt, nil
}
