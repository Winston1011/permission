package main

import (
	"permission/components"
	"permission/conf"
	"permission/helpers"
	"permission/pkg/golib/v2"
	"permission/router"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/server/http"
)

func main() {
	// gin
	engine := gin.New()

	// 初始化基础配置
	helpers.PreInit()
	defer helpers.Clear()

	// ready 探针，支持业务重写
	// base.RegReadyProbe(probe.Ready)
	golib.Bootstraps(engine, golib.BootstrapConf{
		// 业务自定义recover handler
		HandleRecovery: func(c *gin.Context, err interface{}) {
			base.RenderJsonAbort(c, components.ErrorSystemError)
		},
	})
	httpServer(engine)
}

func httpServer(engine *gin.Engine) {
	// web 服务所需资源初始化
	helpers.InitResource(engine)
	defer helpers.Release()

	// 初始化http服务路由
	router.Http(engine)

	// 启动web server
	if err := http.Start(engine, conf.BasicConf.Server); err != nil {
		panic(err.Error())
	}
}
