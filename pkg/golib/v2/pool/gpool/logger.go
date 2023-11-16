package gpool

import (
	"permission/pkg/golib/v2/zlog"
)

// Logger is used for logging formatted messages.
type Logger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
}

var defaultLogger ZLogger

// 为了简单，这里的defaultLogger直接使用zlog了...
type ZLogger struct {
}

func (l ZLogger) Printf(format string, args ...interface{}) {
	zlog.Debugf(nil, format, args...)
}

func (l ZLogger) Print(args ...interface{}) {
	zlog.Debug(nil, args...)
}
