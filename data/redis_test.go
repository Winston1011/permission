package data_test

import (
	"testing"

	"permission/helpers"
)

func TestRedis_Set(t *testing.T) {
	key, value := "k_string_1", "k_string_1"
	// 指定过期时间： PX: 毫秒级
	if _, err := helpers.RedisClient.Do(ctx, "SET", key, value, "PX", 30); err != nil {
		t.Error("[TestRedis_Set] error: ", err.Error())
		return
	}

	// 直接使用 set 设置，expire 使用秒级（EX）
	if err := helpers.RedisClient.Set(ctx, key, key, 30); err != nil {
		t.Error("[TestRedis_Set] error: ", err.Error())
		return
	}
}

func TestRedis_Get(t *testing.T) {
	key := "k_string_1"
	b, err := helpers.RedisClient.Get(ctx, key)
	if err != nil {
		t.Error("[TestRedis_Get] error: ", err.Error())
		return
	}
	t.Logf("[TestRedis_Get] redis get %s return: %s", key, string(b))

	if ttl, err := helpers.RedisClient.Ttl(ctx, key); err == nil {
		t.Log("[TestRedis_Get] ttl: ", ttl)
	}
}

func TestRedis_MSet(t *testing.T) {
	keys := []string{"TestRedis_MSet_MGet_K1", "TestRedis_MSet_MGet_K2"}
	values := []string{"TestRedis_MSet_MGet_V1", "TestRedis_MSet_MGet_V2"}

	if err := helpers.RedisClient.MSet(ctx, keys[0], values[0], keys[1], values[1]); err != nil {
		t.Error("[TestRedis_MSet] error: ", err.Error())
		return
	}

	list := helpers.RedisClient.MGet(ctx, keys...)
	for _, item := range list {
		t.Logf("[TestRedis_MSet] data: %+v", string(item))
	}
}
