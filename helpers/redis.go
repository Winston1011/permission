package helpers

import (
	"permission/conf"

	"permission/pkg/golib/v2/redis"
)

var RedisClient *redis.Redis

// 初始化redis
func InitRedis() {
	c := conf.RConf.Redis["demo"]
	var err error
	RedisClient, err = redis.InitRedisClient(c)
	if err != nil || RedisClient == nil {
		panic("init redis failed!")
	}
}

func CloseRedis() {
	_ = RedisClient.Close()
}
