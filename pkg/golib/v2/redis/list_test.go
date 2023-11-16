package redis

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedis_LPush_LPushX(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()
	key := "TestRedis_LPush"
	list := []interface{}{"1", "2", "3", "4", "5"}

	_, _ = r.Del(ctx, key)

	num, err := r.LPushX(ctx, key, list[0])
	assert.NoError(t, err)
	assert.Equal(t, 0, num)
	n, err := r.LPush(ctx, key, list...)
	assert.NoError(t, err)
	assert.Equal(t, len(list), n)
	time.Sleep(100 * time.Millisecond)
	values, err := r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, len(list), len(values))
	for i := range list {
		assert.Equal(t, list[len(list)-i-1], string(values[i]))
	}

}

func TestRedis_RPush_RPushX(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()
	key := "TestRedis_RPush"
	list := []interface{}{"1", "2", "3", "4", "5"}

	_, _ = r.Del(ctx, key)

	num, err := r.RPushX(ctx, key, list[0])
	assert.NoError(t, err)
	assert.Equal(t, 0, num)

	n, err := r.RPush(ctx, key, list...)
	assert.NoError(t, err)
	assert.Equal(t, len(list), n)

	time.Sleep(100 * time.Millisecond)
	values, err := r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, len(list), len(values))
	for i := range list {
		assert.Equal(t, list[i], string(values[i]))
	}

}

func TestRedis_LPop_RPop_LLen(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()
	key := "TestRedis_LPop_RPop"
	list := []interface{}{"1", "2", "3", "4", "5"}

	_, _ = r.Del(ctx, key)

	_, err := r.RPush(ctx, key, list...)
	assert.NoError(t, err)

	value, err := r.LPop(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, string(value), list[0])

	value, err = r.RPop(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, string(value), list[4])

	time.Sleep(50 * time.Millisecond)
	values, err := r.LLen(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, len(list)-2, values)
}

func TestRedis_LIndex_LSet(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()
	key := "TestRedis_LIndex_LSet"
	list := []interface{}{"1", "2", "3", "4", "5"}

	_, _ = r.Del(ctx, key)

	_, err := r.RPush(ctx, key, list...)
	assert.NoError(t, err)

	value, err := r.LIndex(ctx, key, 0)
	assert.NoError(t, err)
	assert.Equal(t, string(value), list[0])

	ok, err := r.LSet(ctx, key, 0, "SetValue")
	assert.NoError(t, err)
	assert.True(t, ok)

	time.Sleep(50 * time.Millisecond)
	value, err = r.LIndex(ctx, key, 0)
	assert.NoError(t, err)
	assert.Equal(t, "SetValue", string(value))
}

func TestRedis_LRem(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()

	key := "TestRedis_LRem"
	list := []interface{}{"1", "1", "1", "4", "5"}

	_, _ = r.Del(ctx, key)

	value, err := r.LRem(ctx, key, 2, "1")
	assert.NoError(t, err)
	assert.Equal(t, 0, value)

	_, err1 := r.RPush(ctx, key, list...)
	assert.NoError(t, err1)

	value, err = r.LRem(ctx, key, 2, "1")
	assert.NoError(t, err)
	assert.Equal(t, 2, value)

	time.Sleep(50 * time.Millisecond)
	values, err := r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, len(list)-2, len(values))
	for i := range values {
		assert.Equal(t, list[i+2], string(values[i]))
	}

	_, err1 = r.RPush(ctx, key, "1")
	assert.NoError(t, err1)
	value, err = r.LRem(ctx, key, -1, "1")
	assert.NoError(t, err)
	assert.Equal(t, 1, value)

	time.Sleep(50 * time.Millisecond)
	values, err = r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, len(list)-2, len(values))
	for i := range values {
		assert.Equal(t, string(values[i]), list[i+2])
	}

	_, err = r.RPush(ctx, key, "1")
	assert.NoError(t, err)
	value, err = r.LRem(ctx, key, 0, "1")
	assert.NoError(t, err)
	assert.Equal(t, 2, value)

	time.Sleep(50 * time.Millisecond)
	values, err = r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, len(list)-3, len(values))
	for i := range values {
		assert.Equal(t, string(values[i]), list[i+3])
	}
}

func TestRedis_LInsert(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()
	key := "TestRedis_LInsert"
	list := []interface{}{"1", "2", "3", "4", "5"}

	_, _ = r.Del(ctx, key)

	_, err := r.RPush(ctx, key, list...)
	assert.NoError(t, err)

	value, err := r.LInsert(ctx, key, true, "1", "6")
	assert.NoError(t, err)
	assert.Equal(t, len(list)+1, value)

	time.Sleep(50 * time.Millisecond)
	values, err := r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, "6", string(values[0]))

	value, err = r.LInsert(ctx, key, false, "7", "6")
	assert.NoError(t, err)
	assert.Equal(t, -1, value)

	_, _ = r.Del(ctx, key)

	value, err = r.LInsert(ctx, key, false, "1", "6")
	assert.NoError(t, err)
	assert.Equal(t, 0, value)
}

func TestRedis_LTrim(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	setup()
	key := "TestRedis_LTrim"
	list := []interface{}{"1", "2", "3", "4", "5"}

	_, _ = r.Del(ctx, key)

	_, err := r.RPush(ctx, key, list...)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	ok, err := r.LTrim(ctx, key, 1, -1)
	assert.NoError(t, err)
	assert.True(t, ok)

	values, err := r.LRange(ctx, key, 0, 100)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(values))
	list1 := list[1:5]
	for i := range values {
		assert.Equal(t, list1[i], string(values[i]))
	}
}
