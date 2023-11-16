package router

import (
	"time"

	c "permission/controllers/command"
	"permission/helpers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"permission/pkg/golib/v2/command"
)

/*
	命令行一般用在任务中心中启动任务，使用方式(可通过执行 go run main.go -h  查看)：
	goweb application

	Usage:
  		goweb [command]

	Available Commands:
  		job1    This is a job to do xxx
  		job2    This is a job to do yyy

Flags:
  -h, --help   help for goweb


	为了方便，go run main.go 默认启动http服务。
	go run main.go $command  启动一个任务，比如，go run main.go job1
	在任务中心中，执行任务的命令为 bin路径+$command，比如：/usr/local/bin/permission job1
*/
func Commands(rootCmd *cobra.Command, engine *gin.Engine) {
	// 	添加一个名为 job2 的命令，执行方式 go run main.go job2
	var job2Cmd = &cobra.Command{
		Use:   "job1",
		Short: "This is a job to do yyy",
		Run: func(cmd *cobra.Command, args []string) {
			run(engine, c.DemoJob2, args...)
		},
	}
	rootCmd.AddCommand(job2Cmd)

	// 	添加一个名为 job4 的命令，执行方式 go run main.go job4
	var job1Cmd = &cobra.Command{
		Use:   "job2",
		Short: "This is a job to do xxx",
		Run: func(cmd *cobra.Command, args []string) {
			run(engine, c.DemoJob3, args...)
		},
	}
	rootCmd.AddCommand(job1Cmd)
}

func run(engine *gin.Engine, f func(ctx *gin.Context, args ...string) error, args ...string) {
	// 初始化cron任务所需资源（如果任务比较轻量，可单独初始化任务所依赖的资源；否则也可以与webServer共用一个资源初始化方法）
	helpers.InitResourceForCron(engine)

	// 执行任务
	helpers.Job.RunWithRecovery(f, args...)
}

/*
	如果app内需要使用定时任务类，可以通过以下路由加载任务。任务间隔<30min的任务不建议使用上述Commands(任务中心)方式实现。

	crontab与cycle的区别：
	* crontab：每个N时间执行一次，不管上次有没有执行完，N时间后就开始执行下一次任务。
  	比如：N=2min，任务执行了3min。那么程序启动后2分钟执行一次，任务执行了2分后并未结束，但是又开始执行下一次了。
	* cycle：任务执行完后每隔N时间执行一次。
  	比如N=2min，任务执行了3min。程序启动时执行第一次，任务执行完后3+2分后才开始执行第二次任务。
  	需要注意: 除了间隔时间的计算方式不同，第一次执行时间也不同：
	对于每隔N分钟的crontab，服务启动之初会立马执行一次；对于每隔N分钟的cycle，服务启动后N分钟才会执行第一次
*/

func Tasks(engine *gin.Engine) {
	// cycle 任务
	startCycle(engine)
	// 定时任务
	startCrontab(engine)
}
func startCycle(engine *gin.Engine) {
	cycleJob := command.InitCycle(engine)
	cycleJob.AddFunc(time.Second*10, c.DemoJob1)
	cycleJob.Start()
}

func startCrontab(engine *gin.Engine) {
	cronJob := command.InitCrontab(engine)
	_ = cronJob.AddFunc("@every 10s", c.DemoJob4)
	cronJob.Start()
}
