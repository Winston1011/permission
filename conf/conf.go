package conf

import (
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/hbase"
	"permission/pkg/golib/v2/kafka"
	"permission/pkg/golib/v2/redis"
	"permission/pkg/golib/v2/rmq"
	"permission/pkg/golib/v2/server/http"
	"permission/pkg/golib/v2/zlog"
	"permission/pkg/golib/v2/zos"
)

var (
	// 配置文件对应的全局变量
	BasicConf TBasic
	API       TApi
	RConf     ResourceConf
)

// 基础配置,对应config.yaml
type TBasic struct {
	Pprof  base.PprofConfig
	Log    zlog.LogConfig
	Server http.ServerConfig
	// ....业务可扩展其他简单的配置
}

// 对应 api.yaml
type TApi struct {
	Passport base.ApiClient `yaml:"passport"`
}

// 对应 resource.yaml
type ResourceConf struct {
	// Redis    map[string]base.RedisConf // 不建议使用了，改为 redis.RedisConf
	Redis    map[string]redis.RedisConf
	Mysql    map[string]base.MysqlConf
	HBase    map[string]hbase.HBaseConf
	Elastic  map[string]base.ElasticClientConfig
	KafkaPub map[string]kafka.ProducerConfig
	KafkaSub map[string]kafka.ConsumeConfig
	Rmq      rmq.RmqConfig      `yaml:"rmqv2"`
	Zos      zos.CustomerConfig `yaml:"zos"`
}

func InitConf() {
	// 加载通用基础配置（必须）
	env.LoadConf("config.yaml", env.SubConfMount, &BasicConf)

	// 加载api调用相关配置（optional）
	env.LoadConf("api.yaml", env.SubConfMount, &API)

	// 加载资源类配置（optional）
	env.LoadConf("resource.yaml", env.SubConfMount, &RConf)

}
