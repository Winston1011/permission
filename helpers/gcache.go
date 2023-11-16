package helpers

import (
	"time"

	"permission/pkg/golib/v2/gcache"
)

var (
	Cache1 *gcache.BucketCache
	Cache2 *gcache.BucketCache
)

func InitGCache() {
	Cache1 = gcache.NewBucketCache(5*time.Minute, 10*time.Minute, 10)
	Cache2 = gcache.NewBucketCache(45*time.Minute, 1*time.Hour, 10)
}
