package helpers

import (
	"permission/conf"

	"gorm.io/gorm"
	"permission/pkg/golib/v2/base"
)

var (
	MysqlClientPermission *gorm.DB
	MysqlClientWenda      *gorm.DB
	MysqlClientDemo       *gorm.DB
	MysqlClientTest       *gorm.DB
)

func InitMysql() {
	var err error
	for name, dbConf := range conf.RConf.Mysql {
		switch name {
		case "permission":
			MysqlClientPermission, err = base.InitMysqlClient(dbConf)
			//case "wenda":
			//	MysqlClientWenda, err = base.InitMysqlClient(dbConf)
			//case "demo":
			//	MysqlClientDemo, err = base.InitMysqlClient(dbConf)
			//case "test":
			//	MysqlClientTest, err = base.InitMysqlClient(dbConf)
		}

		if err != nil {
			panic("mysql connect error: %v" + err.Error())
		}
	}
}

func CloseMysql() {
	//_ = MysqlClientDemo.Close()
	//_ = MysqlClientTest.Close()
}
