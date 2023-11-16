package helpers

import (
	"time"

	"permission/pkg/golib/v2/base"
)

func InitHttpClient() {
	opt := &base.TransportOption{
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     300 * time.Second,
	}
	base.InitHttp(opt)
}
