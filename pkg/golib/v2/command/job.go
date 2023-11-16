package command

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/job"
	"permission/pkg/golib/v2/middleware"
)

func NewJob(g *gin.Engine) (c *job.Job) {
	c = job.New(g).AddBeforeRun(jobBeforeRun).AddAfterRun(jobAfterRun)
	return c
}

func jobBeforeRun(newCtx *gin.Context, parentContext interface{}) bool {
	middleware.UseMetadata(newCtx)
	middleware.LoggerBeforeRun(newCtx)
	return true
}

func jobAfterRun(ctx *gin.Context) {
	middleware.LoggerAfterRun(ctx)
}
