package command

import (
	"runtime"
	"sync/atomic"
	"time"

	"permission/pkg/golib/v2/pool/gpool"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/zlog"
)

/*
	PoolWithFunc 使用示例： 批量执行同类任务的协程池，每一个 PoolWithFunc 只会绑定一个任务函数，这种 Pool 适用于大批量相同任务的场景。
	因为每个 Pool 只绑定一个任务函数，因此 PoolWithFunc 相较于 Pool 会更加节省内存，但通用性就不如前者
*/

func DemoJob4(ctx *gin.Context) error {
	zlog.Debugf(ctx, "start to do tasks, runtime goroutine is %d", runtime.NumGoroutine())

	runTimes := 10

	// Submit tasks one by one.
	for i := 0; i < runTimes; i++ {
		_ = FuncGoPool.Invoke(i)
	}

	zlog.Debugf(ctx, "running goroutines: %d", FuncGoPool.Running())
	zlog.Debugf(ctx, "finish all tasks, runtime goroutine is %d", runtime.NumGoroutine())

	return nil
}

var FuncGoPool *gpool.PoolWithFunc

func init() {
	// 单独给任务初始化一个连接池，设置连接池的大小为10过期时间为5s

	// 只需要初始化一次!!!
	// 由于该协程池需要传入指定的特定任务，比如这里的 myComplexTask ，所以初始化放到 helpers 里初始化会导致循环引用
	var err error
	FuncGoPool, err = gpool.NewPoolWithFunc(
		10,
		myComplexTask,
		gpool.WithExpiryDuration(5*time.Second),
	)
	if err != nil {
		panic("[NewPoolWithFunc] error: " + err.Error())
	}
}

var sum int32

// 复杂任务示例...
var myComplexTask = func(i interface{}) {
	if n, ok := i.(int); ok {
		atomic.AddInt32(&sum, int32(n))
	}

	zlog.Debugf(nil, "run with %d and sum: %d", i, sum)
}
