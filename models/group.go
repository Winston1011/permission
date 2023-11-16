package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"permission/components"
	"permission/helpers"
)

//添加productid 和 appid 为了防止 权限组同名(针对不同产线)
type Group struct {
	ID         int64   `json:"id" gorm:"primary_key;column:id" `
	ProductID  int64   `json:"productId" gorm:"column:product_id" `
	AppID      int64   `json:"appId" gorm:"column:app_id"`
	GroupName  string  `json:"groupName" gorm:"column:group_name" `
	ParentId   int64   `json:"parentId" gorm:"parent_id"`
	Status     int8    `json:"status" gorm:"column:status" `
	CreateUid  int64   `json:"createUid" gorm:"column:create_uid" `
	UpdateUid  int64   `json:"updateUid" gorm:"column:update_uid" `
	CreateTime int64   `json:"createTime" gorm:"column:create_time" `
	UpdateTime int64   `json:"updateTime" gorm:"column:update_time" `
	Children   []Group `json:"children" gorm:"-"`
}

func (g *Group) TableName() string {
	return components.TABLE_PREX + "group"
}

func (g *Group) InsertGroup(ctx *gin.Context) (err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Create(g).Error
	if err != nil {
		return components.ErrorDbInsert.Wrap(err)
	}
	return nil
}

func (g *Group) BatchInsertGroup(ctx *gin.Context, groups []Group, db *gorm.DB) (rows int64, err error) {
	if len(groups) == 0 {
		return rows, err
	}
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Create(groups)
	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbInsert.Wrap(err)
		return rows, err
	}
	return rows, err
}

func (g *Group) UpsertGroup(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_name", "product_id", "app_id", "status", "update_uid", "update_time"}),
	}).Create(g)
	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbUpsert.Wrap(err)
		return rows, err
	}
	return rows, nil
}

func (g *Group) UpdateGroupById(ctx *gin.Context, id int64, fields map[string]interface{}, db *gorm.DB) (rows int64, err error) {
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Model(&Group{}).Where("`id` = ?", id).Updates(fields)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbUpdate.Wrap(err)
	}
	return rows, nil
}

func (g *Group) GetGroupById(ctx *gin.Context, id int64) (group Group, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where("`id` = ?", id).Take(&group).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return group, components.ErrorDbSelect.Wrap(err)
	}
	return group, nil
}

func (g *Group) GetGroupByConds(ctx *gin.Context, condition map[string]interface{}) (group Group, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Find(&group).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return group, components.ErrorDbSelect.Wrap(err)
	}
	return group, nil
}

func (g *Group) GetGroupListByConds(ctx *gin.Context, condition map[string]interface{}) (groups []Group, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Find(&groups).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return groups, components.ErrorDbSelect.Wrap(err)
	}
	return groups, nil
}

func (g *Group) GetGroupListByPage(ctx *gin.Context, option *Option, page *NormalPage) (groups []Group, cnt int, err error) {
	if !option.IsNeedCnt && !option.IsNeedList {
		return groups, cnt, nil
	}
	db := helpers.MysqlClientPermission.WithContext(ctx).Model(&Group{})
	if option.IsNeedCnt {
		var c int64
		db = db.Count(&c)
		cnt = int(c)
	}
	if option.IsNeedList {
		db = db.Scopes(NormalPaginate(page)).Find(&groups)
	}
	if db.Error != nil {
		return groups, cnt, components.ErrorDbSelect.Wrap(db.Error)
	}
	return groups, cnt, nil
}
