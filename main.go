package main

import (
	"permission/components"
	"permission/conf"
	"permission/helpers"
	"permission/pkg/golib/v2"
	"permission/router"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
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

	/*
		使用cobra来创建CLI命令行
		通过执行 go run main.go -h 可以打印出应用程序支持的命令
		这种方式一般用于项目中有多个执行入口的方式，比如既有http服务又有定时任务或者无web服务但有多个定时任务的情况
		如果项目中只有http服务，可以不使用command方式启动http服务，直接调用http.Start()启动服务即可。
	*/
	var rootCmd = &cobra.Command{
		Use:   "goweb",
		Short: "goweb application ",
		Run: func(cmd *cobra.Command, args []string) {
			// webServer 作为默认命令添加
			// 如果应用只实现定时任务相关，注释掉httpServer，然后在下面的 router.Commands 添加任务
			httpServer(engine)
		},
	}

	// 加载支持的子命令行
	router.Commands(rootCmd, engine)

	if err := rootCmd.Execute(); err != nil {
		panic(err.Error())
	}
}

func httpServer(engine *gin.Engine) {
	// web 服务所需资源初始化
	helpers.InitResource(engine)
	defer helpers.Release()

	// 初始化http服务路由
	router.Http(engine)

	// MQ 消费回调路由
	// router.MQ(engine)

	// app内定时任务（相比任务中心，建议优先使用Tasks方式实现任务）
	// router.Tasks(engine)

	// 启动web server
	if err := http.Start(engine, conf.BasicConf.Server); err != nil {
		panic(err.Error())
	}
}
