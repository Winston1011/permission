package helpers

import (
	"permission/conf"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/zlog"
)

// 基础资源（必须）
func PreInit() {
	// 用于日志中展示模块的名字，开发环境需要手动指定，容器中无需手动指定
	env.SetAppName("permission")

	// 配置加载
	conf.InitConf()

	// 日志初始化  silver_bullet_init_cg_ZfswUVr1mDaLuvrJ8Q
	zlog.InitLog(conf.BasicConf.Log)

}

func maskSensitiveField(fields []zlog.Field) {
	for idx, field := range fields {
		if field.Key == "cell" {
			fields[idx].String = "***"
		}
	}
}

func Clear() {
	// 服务结束时的清理工作，对应 Init() 初始化的资源
	zlog.CloseLogger()
}

// web服务启动所需init的资源
func InitResource(engine *gin.Engine) {
	// 初始化全局变量
	InitMysql()
	InitCasbin()
}

func Release() {
	CloseGPool()
	// CloseKafkaProducer()
	// CloseRocketMq()
}
