// Copyright 2012-present Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"context"
	"log"
	"os"
)

const (
	EsLogKeyStartTime  = "_es_start"
	EsLogKeyEndTime    = "_es_end"
	EsLogKeyStatusCode = "_es_status"
)

// Logger specifies the interface for all log operations.
type Logger interface {
	Printf(ctx context.Context, format string, v ...interface{})
}

type defaultLogger struct {
	*log.Logger
}

func newDefaultLogger(f *os.File) (l *defaultLogger) {
	l.Logger = log.New(f, "", 0)
	return l
}

func (l *defaultLogger) Printf(ctx context.Context, format string, v ...any) {
	l.Logger.Printf(format, v...)
}
