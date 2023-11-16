package command

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/cycle"
	"permission/pkg/golib/v2/middleware"
	"permission/pkg/golib/v2/server/signal"
)

func InitCycle(g *gin.Engine) (c *cycle.Cycle) {
	c = cycle.New(g).AddBeforeRun(cycleBeforeRun).AddAfterRun(cycleAfterRun)
	signal.RegisterShutdown("cycle", c.Stop)
	return c
}

func cycleBeforeRun(ctx *gin.Context) bool {
	middleware.UseMetadata(ctx)
	middleware.LoggerBeforeRun(ctx)
	return true
}

func cycleAfterRun(ctx *gin.Context) {
	middleware.LoggerAfterRun(ctx)
}
