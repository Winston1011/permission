package redis

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRedisSet_GET(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	key, value := "TestRedis_Set_GET_Key", "TestRedis_Set_GET_Value"
	_, err := r.Del(ctx, key)
	assert.NoError(t, err)

	err = r.Set(ctx, key, value)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	data, err := r.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, string(data))
}

func TestRedisSet_Empty(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	key, value := "TestRedis_Set_Empty", ""
	_, err := r.Del(ctx, key)
	assert.NoError(t, err)

	err = r.Set(ctx, key, value)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	data, err := r.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, string(data))
}

func TestRedis_SetEx(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	key, value := "TestRedis_Set_GET_Key", "TestRedis_Set_GET_Value"
	expire := int64(5)
	_, err := r.Del(ctx, key)
	assert.NoError(t, err)

	err = r.SetEx(ctx, key, value, expire)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	ttl, err := r.Ttl(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, expire, ttl)

}

func TestRedis_MSet_MGet(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	keys, values := []string{"TestRedis_MSet_MGet_K1", "TestRedis_MSet_MGet_K2"}, []string{"TestRedis_MSet_MGet_V1", "TestRedis_MSet_MGet_V2"}
	_, _ = r.Del(ctx, keys[0], keys[1])
	time.Sleep(100 * time.Millisecond)
	// MGet 不存在的元素
	data := r.MGet(ctx, keys...)
	assert.Equal(t, "", string(data[0]))
	assert.Equal(t, "", string(data[1]))

	err := r.MSet(ctx, keys[0], values[0], keys[1], values[1])
	assert.NoError(t, err)

	// MGet 存在的元素
	time.Sleep(50 * time.Millisecond)
	data = r.MGet(ctx, keys...)
	assert.NoError(t, err)
	assert.Equal(t, values[0], string(data[0]))
	assert.Equal(t, values[1], string(data[1]))

	// MSet 的传参有误
	err = r.MSet(ctx, keys[0])
	assert.Error(t, err)
}

func TestRedis_Incr_IncrBy_IncrByFloat(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	incrKey := "TestRedis_Incr"
	incrByKey := "TestRedis_IncrBy"
	incrByFloatKey := "TestRedis_IncrByFloat"
	_, _ = r.Del(ctx, incrKey, incrByKey, incrByFloatKey)
	value, err := r.Incr(ctx, incrKey)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), value)

	value, err = r.IncrBy(ctx, incrByKey, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), value)

	fValue, err := r.IncrByFloat(ctx, incrByFloatKey, 1.1)
	assert.NoError(t, err)
	assert.Equal(t, 1.1, fValue)

}
