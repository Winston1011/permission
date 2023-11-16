package golib

import (
	"github.com/gin-gonic/gin"
	_ "go.uber.org/automaxprocs"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/middleware"
	"permission/pkg/golib/v2/utils/gctuner"
)

type BootstrapConf struct {
	Pprof          base.PprofConfig        `yaml:"pprof"`
	AccessLog      middleware.LoggerConfig `yaml:"accessLog"`
	HandleRecovery gin.RecoveryFunc        `yaml:"handleRecovery"`
	GCPercent      uint32                  `yaml:"gcPercent"`
}

func Bootstraps(engine *gin.Engine, conf BootstrapConf) {
	// 环境判断 env GIN_MODE=release/debug
	gin.SetMode(env.RunMode)

	// 过 ingress 的请求clientIp 优先从 "X-Original-Forwarded-For" 中获取
	engine.RemoteIPHeaders = []string{"X-Original-Forwarded-For", "X-Real-IP", "X-Forwarded-For"}

	// Global middleware
	engine.Use(middleware.AccessLog(conf.AccessLog))
	engine.Use(middleware.Recovery(conf.HandleRecovery))
	engine.Use(middleware.Metadata())

	// 就绪探针
	engine.GET("/ready", base.ReadyProbe())

	// gcPercent
	gctuner.Tuning(conf.GCPercent)

	// 性能分析工具
	base.RegisterProf(conf.Pprof)

	// 通用runtime指标采集接口
	base.RegistryMetrics(engine)
}
