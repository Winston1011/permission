package base

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	elasticv7 "github.com/olivere/elastic/v7"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const esPrefix = "@@es."

type ElasticClientConfig struct {
	Addr     string `yaml:"addr"`
	Service  string `yaml:"service"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`

	Sniff       bool `yaml:"sniff"`
	HealthCheck bool `yaml:"healthCheck"`
	Gzip        bool `yaml:"gzip"`

	DebugMsgLen int `yaml:"debugMsgLen"`
	InfoMsgLen  int `yaml:"infoMsgLen"`

	Decoder       elastic.Decoder
	RetryStrategy elastic.Retrier
	HttpClient    *http.Client
}

func (conf *ElasticClientConfig) checkConfig() {
	env.CommonSecretChange(esPrefix, *conf, conf)
}

func NewESClientV7(cfg ElasticClientConfig, others ...elasticv7.ClientOptionFunc) (*elasticv7.Client, error) {
	cfg.checkConfig()

	addrs := strings.Split(cfg.Addr, ",")
	options := []elasticv7.ClientOptionFunc{
		elasticv7.SetURL(addrs...),
		elasticv7.SetSniff(cfg.Sniff),
		elasticv7.SetHealthcheck(cfg.HealthCheck),
		elasticv7.SetGzip(cfg.Gzip),
	}

	logger := zlog.ZapLogger.WithOptions(zlog.AddCallerSkip(1))
	options = append(
		options,
		elasticv7.SetTraceLog(esLogger{logger: logger, addr: cfg.Addr, service: cfg.Service, level: "trace", msgLen: cfg.DebugMsgLen}),
		elasticv7.SetInfoLog(esLogger{logger: logger, addr: cfg.Addr, service: cfg.Service, level: "info"}),
		elasticv7.SetErrorLog(esLogger{logger: logger, addr: cfg.Addr, service: cfg.Service, level: "error"}),
	)

	if cfg.Username != "" || cfg.Password != "" {
		options = append(options, elasticv7.SetBasicAuth(cfg.Username, cfg.Password))
	}

	if cfg.HttpClient != nil {
		options = append(options, elasticv7.SetHttpClient(cfg.HttpClient))
	}
	if cfg.Decoder != nil {
		options = append(options, elasticv7.SetDecoder(cfg.Decoder))
	}

	if cfg.RetryStrategy != nil {
		options = append(options, elasticv7.SetRetrier(cfg.RetryStrategy))
	}

	// override
	if len(others) > 0 {
		options = append(options, others...)
	}
	return elasticv7.NewClient(options...)
}

type esLogger struct {
	logger  *zlog.Logger
	addr    string
	service string
	level   string
	msgLen  int
}

func (l esLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	// trace 日志允许通过 msgLen 字段控制是否打印
	if l.level == "trace" && l.msgLen == -1 {
		// 不打印trace日志
		return
	}

	// msg 字段截断(目前仅支持trace日志截断)
	msg := fmt.Sprintf(format, v...)
	if l.msgLen > 0 && len(msg) > l.msgLen {
		msg = msg[:l.msgLen]
	}

	// 通用字段
	var logID, requestID, uri string
	// info 日志特有字段
	var start, end time.Time
	var httpCode int
	if c, ok := ctx.(*gin.Context); (ok && c != nil) || (!ok && !utils.IsNil(ctx)) {
		logID, _ = ctx.Value(zlog.ContextKeyLogID).(string)
		requestID, _ = ctx.Value(zlog.ContextKeyRequestID).(string)
		uri, _ = ctx.Value(zlog.ContextKeyUri).(string)

		if l.level == "info" {
			start, _ = ctx.Value(elasticv7.EsLogKeyStartTime).(time.Time)
			end, _ = ctx.Value(elasticv7.EsLogKeyEndTime).(time.Time)
			httpCode, _ = ctx.Value(elasticv7.EsLogKeyStatusCode).(int)
		}
	}

	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("prot", "es"),
		zlog.String("service", l.service),
		zlog.String("addr", l.addr),
		zlog.String("localIp", env.LocalIP),
		zlog.String("module", env.GetAppName()),
		zlog.String("requestId", requestID),
		zlog.String("logID", logID),
		zlog.String("uri", uri),
	}

	switch l.level {
	case "trace":
		l.logger.Debug(msg, fields...)
	case "error":
		l.logger.Error(msg, fields...)
	case "info":
		ralCode := -1
		if httpCode == http.StatusOK {
			ralCode = 0
		}

		fields = append(fields,
			zlog.Int("httpCode", httpCode),
			zlog.Int("ralCode", ralCode),
			zlog.Float64("cost", utils.GetRequestCost(start, end)),
			zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
			zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		)
		l.logger.Info(msg, fields...)
	}
}
