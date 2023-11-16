package helpers

import (
	"permission/conf"

	"permission/pkg/golib/v2/hbase"
)

var HBaseDemo *hbase.HBaseClient

func InitHBase() {
	demoHBaseConf := conf.RConf.HBase["demo"]
	var err error
	HBaseDemo, err = hbase.NewHBaseClient(demoHBaseConf)
	if err != nil {
		panic("init hbase client error: " + err.Error())
	}
	if HBaseDemo == nil {
		panic("init hbase pool failed")
	}
}

func CloseHBase() {
	_ = HBaseDemo.Close()
}
