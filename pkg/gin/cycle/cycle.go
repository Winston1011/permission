package cycle

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Cycle struct {
	entries   []*Entry
	gin       *gin.Engine
	beforeRun func(*gin.Context) bool
	afterRun  func(*gin.Context)
	stop      bool
}

type Job interface {
	Run(ctx *gin.Context) error
}

type Entry struct {
	Duration time.Duration
	Job      Job
}

func New(engine *gin.Engine) *Cycle {
	return &Cycle{
		entries: nil,
		gin:     engine,
	}
}

type FuncJob func(*gin.Context) error

func (f FuncJob) Run(ctx *gin.Context) error { return f(ctx) }

// add cron before func
func (c *Cycle) AddBeforeRun(beforeRun func(*gin.Context) bool) *Cycle {
	c.beforeRun = beforeRun
	return c
}

// add cron after func
func (c *Cycle) AddAfterRun(afterRun func(*gin.Context)) *Cycle {
	c.afterRun = afterRun
	return c
}

func (c *Cycle) AddFunc(duration time.Duration, cmd func(*gin.Context) error) {
	entry := &Entry{
		Duration: duration,
		Job:      FuncJob(cmd),
	}
	c.entries = append(c.entries, entry)
}

func (c *Cycle) Start() {
	for _, e := range c.entries {
		go c.run(e)
	}
}

func (c *Cycle) Stop(ctx context.Context) error {
	c.stop = true
	return nil
}

// 死循环
func (c *Cycle) run(e *Entry) {
	for {
		c.runWithRecovery(e)

		if c.stop {
			return
		}
	}
}

func (c *Cycle) runWithRecovery(entry *Entry) {
	ctx := gin.CreateNewContext(c.gin)
	ctx.CustomContext.Handle = entry.Job.(FuncJob)
	ctx.CustomContext.Desc = string(entry.Duration)
	ctx.CustomContext.Type = "Cycle"
	ctx.CustomContext.StartTime = time.Now()

	defer func() {
		if r := recover(); r != nil {
			handleName := ctx.CustomContext.HandlerName()
			gin.CustomerErrorLog(ctx, fmt.Sprintf("cycle panic: %v", r), true, map[string]string{
				"handle": handleName,
			})
		}

		gin.RecycleContext(c.gin, ctx)

		time.Sleep(entry.Duration)
	}()

	if c.beforeRun != nil {
		ok := c.beforeRun(ctx)
		if !ok {
			return
		}
	}

	err := entry.Job.Run(ctx)
	ctx.CustomContext.Error = err
	ctx.CustomContext.EndTime = time.Now()

	if c.afterRun != nil {
		c.afterRun(ctx)
	}
}
