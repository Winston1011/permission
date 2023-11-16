package helpers

import (
	"runtime"

	"permission/pkg/golib/v2/pool/gpool"
	"permission/pkg/golib/v2/zlog"
)

// 全局共用的协程池
var GoPool *gpool.Pool

func InitGPool() {
	var err error
	GoPool, err = gpool.NewPool(100)
	if err != nil {
		panic("[initGPool error: " + err.Error())
	}
}

func CloseGPool() {
	if GoPool != nil {
		GoPool.Release()
		zlog.Debugf(nil, "exit, runtime goroutine is %d", runtime.NumGoroutine())
	}
}
