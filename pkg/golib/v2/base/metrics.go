package base

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"permission/pkg/golib/v2/zlog"
)

var RuntimeMetricsRegister *prometheus.Registry

func RegistryMetrics(engine *gin.Engine) {
	prometheus.DefaultRegisterer.Unregister(collectors.NewGoCollector())
	prometheus.DefaultRegisterer.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	RuntimeMetricsRegister = prometheus.NewRegistry()
	RuntimeMetricsRegister.MustRegister(collectors.NewGoCollector(collectors.WithGoCollections(collectors.GoRuntimeMetricsCollection)))

	engine.GET("/runtime-metrics", func(ctx *gin.Context) {
		// 避免metrics打点输出过多无用日志
		zlog.SetNoLogFlag(ctx)

		httpHandler := promhttp.InstrumentMetricHandler(
			RuntimeMetricsRegister, promhttp.HandlerFor(RuntimeMetricsRegister, promhttp.HandlerOpts{}),
		)
		httpHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})
}
