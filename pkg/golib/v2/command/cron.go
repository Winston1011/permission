package command

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/cron"
	"permission/pkg/golib/v2/middleware"
	"permission/pkg/golib/v2/server/signal"
)

func InitCrontab(g *gin.Engine) (c *cron.Cron) {
	c = cron.New(
		g,
		cron.WithSeconds(),
		cron.WithChain(
			cron.Recover(),
		)).
		AddBeforeRun(cronBeforeRun).
		AddAfterRun(cronAfterRun)

	signal.RegisterShutdown("crontab", c.Shutdown)

	c.Start()
	return c
}

func cronBeforeRun(ctx *gin.Context) bool {
	middleware.UseMetadata(ctx)
	middleware.LoggerBeforeRun(ctx)
	return true
}

func cronAfterRun(ctx *gin.Context) {
	middleware.LoggerAfterRun(ctx)
}
