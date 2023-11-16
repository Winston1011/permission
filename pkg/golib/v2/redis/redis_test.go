package redis

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var r *Redis

func init() {
	setup()
}

func setup() {
	// 初始化redis
	demoRedisConf := RedisConf{
		Service:      "demo",
		Addr:         "127.0.0.1:6379",
		Password:     "",
		MaxIdle:      10,
		MaxActive:    20,
		IdleTimeout:  600 * time.Second,
		ConnTimeOut:  1 * time.Second,
		ReadTimeOut:  1 * time.Second,
		WriteTimeOut: 1 * time.Second,
	}
	objRedis, _ := InitRedisClient(demoRedisConf)
	if objRedis == nil {
		panic("init redis failed!")
	}
	r = objRedis

}

func TestDo(t *testing.T) {
	setup()
	t.Run("Do", func(t *testing.T) {
		reply, err := r.Do(nil, "ping")
		assert.Equal(t, "PONG", reply)
		assert.Equal(t, nil, err)
	})
}
