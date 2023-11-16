package job

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Job struct {
	gin           *gin.Engine
	beforeRun     func(*gin.Context, interface{}) bool
	afterRun      func(*gin.Context)
	getJobContext func(*gin.Context) interface{}
}

type Entry struct {
	Job FuncJob
}

type FuncJob func(*gin.Context) error

func (f FuncJob) Run(ctx *gin.Context) error { return f(ctx) }

func New(engine *gin.Engine) *Job {
	return &Job{
		gin: engine,
	}
}

// add cron before func
func (c *Job) AddBeforeRun(beforeRun func(*gin.Context, interface{}) bool) *Job {
	c.beforeRun = beforeRun
	return c
}

// add cron after func
func (c *Job) AddAfterRun(afterRun func(*gin.Context)) *Job {
	c.afterRun = afterRun
	return c
}

func (c *Job) AddJobContext(jobContext func(*gin.Context) interface{}) *Job {
	c.getJobContext = jobContext
	return c
}

func (c *Job) Run(ctx *gin.Context, f func(ctx *gin.Context) error) {
	go c.recoverRun(f)
}

func (c *Job) RunSync(ctx *gin.Context, f func(ctx *gin.Context) error) {
	c.recoverRun(f)
}

func (c *Job) recoverRun(f FuncJob) {
	ctx := gin.CreateNewContext(c.gin)
	ctx.CustomContext.Handle = f
	ctx.CustomContext.Type = "Job"
	ctx.CustomContext.StartTime = time.Now()

	defer c.runEnd(ctx)

	if c.beforeRun != nil {
		_ = c.beforeRun(ctx, "")
	}

	err := f.Run(ctx)
	ctx.CustomContext.Error = err
	ctx.CustomContext.EndTime = time.Now()

	if c.afterRun != nil {
		c.afterRun(ctx)
	}
}

func (c *Job) runEnd(ctx *gin.Context) {
	if r := recover(); r != nil {
		gin.CustomerErrorLog(ctx, fmt.Sprintf("job panic: %v", r), true, map[string]string{
			"desc": ctx.CustomContext.Desc,
		})
	}
	gin.RecycleContext(c.gin, ctx)
}

func (c *Job) RunWithRecovery(f func(*gin.Context, ...string) error, args ...string) {
	ctx := gin.CreateNewContext(c.gin)

	h := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	ctx.CustomContext.Desc = "handler: " + h + ", args:" + strings.Join(args, ", ")
	ctx.CustomContext.Type = "Job"
	ctx.CustomContext.StartTime = time.Now()

	defer c.runEnd(ctx)

	if c.beforeRun != nil {
		c.beforeRun(ctx, nil)
	}

	err := f(ctx, args...)

	ctx.CustomContext.Error = err
	ctx.CustomContext.EndTime = time.Now()

	if c.afterRun != nil {
		c.afterRun(ctx)
	}
}
