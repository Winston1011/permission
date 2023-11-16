package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"permission/components"
	"permission/helpers"
)

type GroupNode struct {
	ID       int64 `json:"id" gorm:"primary_key;column:id" `
	GroupId  int64 `json:"groupId" gorm:"column:group_id" `
	NodeId   int64 `json:"nodeId" gorm:"column:node_id" `
	NodeType int8  `json:"nodeType" gorm:"column:node_type"`
}

func (gn *GroupNode) TableName() string {
	return components.TABLE_PREX + "rel_group_node"
}

func (gn *GroupNode) InsertGroupNode(ctx *gin.Context) (err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Create(gn).Error
	if err != nil {
		return components.ErrorDbInsert.Wrap(err)
	}
	return nil
}

func (gn *GroupNode) BatchInsertGroupNode(ctx *gin.Context, groupNodes []GroupNode, db *gorm.DB) (rows int64, err error) {
	if len(groupNodes) == 0 {
		return rows, err
	}
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Create(groupNodes)
	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbInsert.Wrap(err)
		return rows, err
	}
	return rows, err
}

func (gn *GroupNode) UpsertGroupNode(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_id", "node_id"}),
	}).Create(gn)
	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbUpsert.Wrap(err)
		return rows, err
	}
	return rows, nil
}

func (gn *GroupNode) UpdateGroupNodeById(ctx *gin.Context, fields map[string]interface{}) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Model(gn).Updates(fields)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbUpdate.Wrap(err)
	}
	return rows, nil
}

func (gn *GroupNode) BatchDeleteGroupNode(ctx *gin.Context, nodeIds []int64, db *gorm.DB) (rows int64, err error) {
	if len(nodeIds) == 0 {
		return rows, err
	}
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).
		Where("group_id = ?", gn.GroupId).
		Where("node_id IN ?", nodeIds).
		Delete(GroupNode{})
	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbDelete.Wrap(err)
		return rows, err
	}
	return rows, err
}

func (gn *GroupNode) GetGroupNodeById(ctx *gin.Context, id int64) (groupNode GroupNode, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where("`id` = ?", id).Take(&groupNode).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return groupNode, components.ErrorDbSelect.Wrap(err)
	}
	return groupNode, nil
}

func (gn *GroupNode) GetGroupNodeListByConds(ctx *gin.Context, condition map[string]interface{}) (groupNodes []GroupNode, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Find(&groupNodes).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return groupNodes, components.ErrorDbSelect.Wrap(err)
}

func (gn *GroupNode) GetGroupNodeListByPage(ctx *gin.Context, option *Option, page *NormalPage) (groupNodeList []GroupNode, cnt int, err error) {
	if !option.IsNeedCnt && !option.IsNeedList {
		return groupNodeList, cnt, nil
	}
	db := helpers.MysqlClientPermission.WithContext(ctx).Model(&GroupNode{})
	if option.IsNeedCnt {
		var c int64
		db = db.Count(&c)
		cnt = int(c)
	}
	if option.IsNeedList {
		db = db.Scopes(NormalPaginate(page)).Find(&groupNodeList)
	}
	if db.Error != nil {
		return groupNodeList, cnt, components.ErrorDbSelect.Wrap(err)
	}
	return groupNodeList, cnt, nil
}
