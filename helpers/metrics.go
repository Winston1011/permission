package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"permission/pkg/golib/v2/zlog"
)

var HttpCounterVec *prometheus.CounterVec
var HttpReqDurationsHistogram *prometheus.HistogramVec

func InitPromMetrics(router *gin.Engine) {
	metricsRegister := prometheus.NewRegistry()
	HttpCounterVec = prometheus.NewCounterVec(
		// Namespace, Subsystem, and Name are components of the fully-qualified
		// name of the Metric (created by joining these components with
		// "_"). Only Name is mandatory, the others merely help structuring the
		// name. Note that the fully-qualified name of the metric must be a
		// valid Prometheus metric name.
		prometheus.CounterOpts{
			// 建议以业务线的名字进行命名
			Namespace: "pkg",
			// 建议使用模块名进行命名
			Subsystem: "goweb",
			// 指标名称，全名就是 inf_godemo_http_request_total ,通过 namespace 和 subsystem 保证了指标名称的唯一
			Name: "http_request_total",
			Help: "total number of v1.getUserInfo",
		},
		[]string{"code", "requestType", "handler"},
	)

	HttpReqDurationsHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pkg",
			Subsystem: "goweb",
			Name:      "http_request_duration_millisecond",
			Help:      "http request durations in millisecond.",
			// 4 buckets, starting from 0.1 and adding 0.5 between each bucket
			Buckets: prometheus.LinearBuckets(0.1, 0.5, 4),
		},
		[]string{"method", "path"},
	)

	// 注册指标
	if err := metricsRegister.Register(HttpCounterVec); err != nil {
		// 自行处理错误（一般注册失败会panic，避免后面采集受影响）
		panic("register error: " + err.Error())
	}
	// 注册失败会panic：
	metricsRegister.MustRegister(HttpReqDurationsHistogram)

	// 	下面方法对外暴露了 /metrics 的采集接口，给出了两种示例:

	// 1. 采集接口uri默认为 /metrics ，可以自行修改
	router.GET("/metrics", func(ctx *gin.Context) {
		// 避免metrics打点输出过多无用日志
		zlog.SetNoLogFlag(ctx)
		httpHandler := promhttp.InstrumentMetricHandler(
			metricsRegister, promhttp.HandlerFor(metricsRegister, promhttp.HandlerOpts{}),
		)

		httpHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	/*
		2. 之前php sdk 中封装的方法，默认采集后会主动flush内存（清空指标数据）,
		原因：如果采集后不重置指标，那么应用程序重启后的一次采集会与上一次采集指标变化较大
		所以php 迁移go后，用户反馈需要提供该种采集方式。
		可以通过重写 promhttp.Handler() 实现，参见promServer方法：
	*/
	// router.GET("/metrics", promServer)
}

func promServer(ctx *gin.Context) {
	zlog.SetNoLogFlag(ctx)

	// 这个handler是sdk中提供的默认的handler
	// 如果只是想重置指标，可以保持原样;
	// 如果想要实现自己的handler，可以修改

	promHandler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer, promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{}),
	)

	promHandler.ServeHTTP(ctx.Writer, ctx.Request)

	// 重置指标
	HttpCounterVec.Reset()
	HttpReqDurationsHistogram.Reset()
}
