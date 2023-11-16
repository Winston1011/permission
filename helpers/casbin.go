package helpers

import (
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

const modelPolicyAddr = "conf/rbac_model.conf"

var (
	Enforcer *casbin.Enforcer
	Adapter  *gormadapter.Adapter
)

type CasbinRule struct {
	ID    int64  `gorm:"primaryKey;autoIncrement;not null"`
	Ptype string `gorm:"v1"`
	V0    string `gorm:"column:v0"`
	V1    string `gorm:"column:v1"`
	V2    string `gorm:"column:v2"`
	V3    string `gorm:"column:v3"`
	V4    string `gorm:"column:v4"`
	V5    string `gorm:"v5" default:""`
}

func InitCasbin() {
	Adapter, _ = gormadapter.NewAdapterByDBWithCustomTable(MysqlClientPermission, CasbinRule{}, "tb_permission_casbin_rule")
	Enforcer, _ = casbin.NewEnforcer(modelPolicyAddr, Adapter)
	Enforcer.LoadModel()
	Enforcer.LoadPolicy()
}
