package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"permission/components"
	"permission/helpers"
)

type CasbinRule struct {
	ID              int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Ptype           string `json:"policyType"`                      // "p"
	GroupId         string `json:"groupId" gorm:"column:v0"`        // 2
	ProductAppField string `json:"domain" gorm:"column:v1"`         // 111:222
	Resource        string `json:"resource" gorm:"column:v2"`       // furi:uri
	PermissionType  string `json:"permissionType" gorm:"column:v3"` // read or write (当先统一为any， 后续根据具体进行划分)
	Status          string `json:"status" gorm:"column:v4"`         //allow/deny
	V5              string `gorm:"v5" default:""`
}

func (cr *CasbinRule) TableName() string {
	return components.TABLE_PREX + "casbin_rule"
}

func (cr *CasbinRule) InsertCasbinRule(ctx *gin.Context) (err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Create(cr).Error
	//cr.ID 可以获得数据库插入时获得的自增ID的值
	//id := cr.ID
	if err != nil {
		return components.ErrorDbInsert.Wrap(err)
	}
	return nil
}

func (cr *CasbinRule) BatchInsertCasbinRule(ctx *gin.Context, crs []CasbinRule, db *gorm.DB) (rows int64, err error) {
	if len(crs) == 0 {
		return rows, err
	}
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Create(crs)
	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbInsert.Wrap(err)
		return rows, err
	}
	return rows, err
}

func (cr *CasbinRule) BatchUpsertCasbinRule(ctx *gin.Context, crs []CasbinRule, db *gorm.DB) (rows int64, err error) {
	if len(crs) == 0 {
		return rows, err
	}
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(crs)
	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbUpsert.Wrap(err)
	}
	return rows, err
}

func (cr *CasbinRule) UpsertCasbinRule(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"ptype", "v0", "v1", "v2", "v3", "v4", "v5"}),
	}).Create(cr)
	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbUpsert.Wrap(err)
		return rows, err
	}
	return rows, nil
}

func (cr *CasbinRule) UpdateCasbinRuleById(ctx *gin.Context, id int64, fields map[string]interface{}) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Model(cr).Where("`id` = ?", id).Updates(fields)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbUpdate.Wrap(err)
	}
	return rows, nil
}

func (cr *CasbinRule) DeleteCasbinRule(ctx *gin.Context) (rows int64, err error) {
	db := helpers.MysqlClientPermission
	result := db.WithContext(ctx).Delete(cr)
	rows, err = result.RowsAffected, result.Error
	if err != nil {
		return rows, components.ErrorDbDelete.Wrap(err)
	}
	return rows, nil
}

func (cr *CasbinRule) DeleteCasbinRuleByCondition(ctx *gin.Context, condition map[string]interface{}, db *gorm.DB) (rows int64, err error) {
	if db == nil {
		db = helpers.MysqlClientPermission
	}
	result := db.WithContext(ctx).Where(condition).Delete(CasbinRule{})
	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbDelete.Wrap(err)
		return rows, err
	}
	return rows, err
}

func (cr *CasbinRule) GetCasbinRuleById(ctx *gin.Context, id int64) (rule CasbinRule, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where("`id` = ?", id).Take(&rule).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return rule, components.ErrorDbSelect.Wrap(err)
	}
	return rule, nil
}

func (cr *CasbinRule) GetCasbinRulesListByConds(ctx *gin.Context, condition map[string]interface{}) (rules []CasbinRule, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Find(&rules).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return rules, components.ErrorDbSelect.Wrap(err)
	}
	return rules, nil
}

func (cr *CasbinRule) GetCasbinRulesByConds(ctx *gin.Context, condition map[string]interface{}) (rule CasbinRule, err error) {
	db := helpers.MysqlClientPermission
	err = db.WithContext(ctx).Where(condition).Find(&rule).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return rule, components.ErrorDbSelect.Wrap(err)
	}
	return rule, nil
}

func (cr *CasbinRule) GetCasbinRulesByPage(ctx *gin.Context, option *Option, page *NormalPage) (rules []CasbinRule, cnt int, err error) {
	if !option.IsNeedCnt && !option.IsNeedList {
		return rules, cnt, nil
	}
	db := helpers.MysqlClientPermission.WithContext(ctx).Model(&CasbinRule{})
	if option.IsNeedCnt {
		var c int64
		db = db.Count(&c)
		cnt = int(c)
	}
	if option.IsNeedList {
		db = db.Scopes(NormalPaginate(page)).Find(&rules)
	}
	if db.Error != nil {
		return rules, cnt, components.ErrorDbSelect.Wrap(db.Error)
	}
	return rules, cnt, nil
}
