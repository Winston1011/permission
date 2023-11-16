package command

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"permission/helpers"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/zlog"
)

// 一个goroutine pool的使用示例
func DemoJob3(ctx *gin.Context, args ...string) error {
	zlog.Debugf(ctx, "start to do tasks, runtime goroutine is %d", runtime.NumGoroutine())

	runTimes := 1000

	var wg sync.WaitGroup
	for i := 0; i < runTimes; i++ {
		wg.Add(1)

		idx := i
		newCtx := ctx.Copy()
		syncCalculateSum := func() {
			defer wg.Done()
			demoFunc(newCtx, idx)
		}

		// 使用全局通用连接池
		_ = helpers.GoPool.Submit(syncCalculateSum)
	}
	wg.Wait()

	zlog.Debugf(ctx, "running goroutines: %d", helpers.GoPool.Running())
	zlog.Debugf(ctx, "finish all tasks, runtime goroutine is %d", runtime.NumGoroutine())

	time.Sleep(5 * time.Second)

	return nil
}

func demoFunc(ctx *gin.Context, runTimes int) {
	time.Sleep(10 * time.Millisecond)

	k := fmt.Sprintf("K_%d", runTimes)
	zlog.Debug(ctx, "Hello ", k)

	zlog.AddField(ctx, zlog.Int(k, runTimes))
}
