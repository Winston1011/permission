package cron

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// JobWrapper decorates the given Job with some behavior.
type JobWrapper func(Job) Job

// Chain is a sequence of JobWrappers that decorates submitted jobs with
// cross-cutting behaviors like logging or synchronization.
type Chain struct {
	wrappers []JobWrapper
}

// NewChain returns a Chain consisting of the given JobWrappers.
func NewChain(c ...JobWrapper) Chain {
	return Chain{c}
}

// Then decorates the given job with all JobWrappers in the chain.
//
// This:
//     NewChain(m1, m2, m3).Then(job)
// is equivalent to:
//     m1(m2(m3(job)))
func (c Chain) Then(j Job) Job {
	for i := range c.wrappers {
		j = c.wrappers[len(c.wrappers)-i-1](j)
	}
	return j
}

// Recover panics in wrapped jobs and log them with the provided logger.
func Recover() JobWrapper {
	return func(j Job) Job {
		return FuncJob(func(ctx *gin.Context) error {

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("cron panic: %v", r)
					}
					gin.CustomerErrorLog(ctx, err.Error(), true, nil)
				}
			}()

			return j.Run(ctx)
		})
	}
}

// DelayIfStillRunning serializes jobs, delaying subsequent runs until the
// previous one is complete. Jobs running after a delay of more than a minute
// have the delay logged at Info.
func DelayIfStillRunning() JobWrapper {
	return func(j Job) Job {
		var mu sync.Mutex
		return FuncJob(func(ctx *gin.Context) error {
			start := time.Now()
			mu.Lock()
			defer mu.Unlock()
			if dur := time.Since(start); dur > time.Minute {
				gin.DefaultLogger.Print("delay , duration: " + dur.String())
			}
			return j.Run(ctx)
		})
	}
}

// SkipIfStillRunning skips an invocation of the Job if a previous invocation is
// still running. It logs skips to the given logger at Info level.
func SkipIfStillRunning() JobWrapper {
	return func(j Job) Job {
		var ch = make(chan struct{}, 1)
		ch <- struct{}{}
		return FuncJob(func(ctx *gin.Context) (err error) {
			select {
			case v := <-ch:
				defer func() { ch <- v }()
				err = j.Run(ctx)
			default:
				gin.DefaultLogger.Print("skip")
			}
			return err
		})
	}
}
