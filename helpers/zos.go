package helpers

import (
	"permission/conf"

	"permission/pkg/golib/v2/zos"
)

var Bucket zos.Bucket

func InitZos() {
	bucketList := zos.NewBucket(conf.RConf.Zos)
	// 注意这里的 epoch-inf-callcenter 是从 OP老师 那申请的 bucket
	Bucket = bucketList["epoch-ship"]
}
