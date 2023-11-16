package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"permission/components"
	"permission/helpers"
)

type Node struct {
	ID         int64  `json:"id" gorm:"primary_key;column:id" `
	ProductID  int64  `json:"productId" gorm:"column:product_id" `
	AppID      int64  `json:"appId" gorm:"column:app_id" `
	Label      string `json:"label" gorm:"column:label" `
	Resource   string `json:"resource" gorm:"column:resource" `
	NodeType   int8   `json:"nodeType" gorm:"column:node_type"`
	IsShow     int8   `json:"isShow" gorm:"column:is_show"`
	ParentID   int64  `json:"parentId" gorm:"column:parent_id" `
	CreateUid  int64  `json:"createUid" gorm:"column:create_uid" `
	UpdateUid  int64  `json:"updateUid" gorm:"column:update_uid" `
	CreateTime int64  `json:"createTime" gorm:"column:create_time" `
	UpdateTime int64  `json:"updateTime" gorm:"column:update_time" `
	Children   []Node `json:"children" gorm:"-"`
	Selected   int8   `json:"selected" gorm:"-"`
}

func (n *Node) TableName() string {
	return components.TABLE_PREX + "node"
}

func (n *Node) InsertNode(ctx *gin.Context) (err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Create(n).Error
	if err != nil {
		return components.ErrorDbInsert.Wrap(err)
	}
	return nil
}

func (n *Node) BatchInsertNode(ctx *gin.Context, Nodes []Node, db *gorm.DB) (rows int64, err error) {
	if len(Nodes) == 0 {
		return rows, err
	}
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Create(Nodes)
	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbInsert.Wrap(err)
		return rows, err
	}
	return rows, err
}

func (n *Node) UpsertNode(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "desc"}),
	}).Create(n)
	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbUpsert.Wrap(err)
		return rows, err
	}
	return rows, nil
}

func (n *Node) UpdateNodeById(ctx *gin.Context, id int64, fields map[string]interface{}) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Model(&Node{}).Where("`id` = ?", id).Updates(fields)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbUpdate.Wrap(err)
	}
	return rows, nil
}

func (n *Node) DeleteNodeById(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Delete(n)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbDelete.Wrap(err)
	}
	return rows, nil
}

func (n *Node) GetNodeById(ctx *gin.Context, id int64) (node Node, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where("`id` = ?", id).Take(&node).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return node, components.ErrorDbSelect.Wrap(err)
	}
	return node, nil
}

func (n *Node) GetNodeListByCondition(ctx *gin.Context, condition map[string]interface{}) (nodes []Node, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Order("id").Find(&nodes).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return nodes, components.ErrorDbSelect.Wrap(err)
	}
	return nodes, nil
}

func (n *Node) GetNodeByCondition(ctx *gin.Context, condition map[string]interface{}) (node Node, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Find(&node).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return node, components.ErrorDbSelect.Wrap(err)
	}
	return node, nil
}

func (n *Node) GetNodeListByPage(ctx *gin.Context, option *Option, page *NormalPage) (nodes []Node, cnt int, err error) {
	if !option.IsNeedCnt && !option.IsNeedList {
		return nodes, cnt, nil
	}
	db := helpers.MysqlClientPermission.WithContext(ctx).Model(&Node{})
	if option.IsNeedCnt {
		var c int64
		db = db.Count(&c)
		cnt = int(c)
	}
	if option.IsNeedList {
		db = db.Scopes(NormalPaginate(page)).Find(&nodes)
	}
	if db.Error != nil {
		return nodes, cnt, components.ErrorDbSelect.Wrap(db.Error)
	}
	return nodes, cnt, nil
}
