package base

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"
	"permission/pkg/golib/v2/env"
)

type PprofConfig struct {
	Enable bool   `yaml:"enable"`
	Port   string `yaml:"port"`
}

// 为方便采集任务处理，线上环境不再支持自定义端口和开关。默认开启，并使用8060端口。
func RegisterProf(conf PprofConfig) {
	port := ":8060"
	if env.GetRunEnv() == env.RunEnvTest {
		// 开发环境允许自定义开关及端口
		if !conf.Enable {
			return
		}
		if conf.Port != "" {
			port = conf.Port
		}
	}

	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())

	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			panic("go profiler server start error: " + err.Error())
		}
	}()
}
