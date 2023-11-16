package redis

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRedis_SAdd_SIsMember_SMembers(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	key, values := "TestRedis_SAdd_SIsMember_SMembers", []string{"1", "2", "3", "3"}
	_, _ = r.Del(ctx, key)

	n, err := r.SAdd(ctx, key, values...)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(values))-1, n)

	time.Sleep(50 * time.Millisecond)
	ok, err := r.SIsMember(ctx, key, values[0])
	assert.NoError(t, err)
	assert.True(t, ok)

	members, err := r.SMembers(ctx, key)
	assert.NoError(t, err)
	t.Log("members len :", len(members))
	assert.Equal(t, len(values)-1, len(members))
}

func TestRedis_SRem_SCard(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	key, values := "TestRedis_SAdd_SIsMember_SMembers", []string{"1", "2", "3", "3"}

	_, _ = r.Del(ctx, key)

	n, err := r.SAdd(ctx, key, values...)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(values))-1, n)

	time.Sleep(50 * time.Millisecond)
	ok, err := r.SIsMember(ctx, key, values[0])
	assert.NoError(t, err)
	assert.True(t, ok)

	members, err := r.SMembers(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, len(values)-1, len(members))
}

func TestRedis_SPop_SRandMember(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	key, values := "TestRedis_SAdd_SIsMember_SMembers", []string{"1", "2", "3"}

	_, _ = r.Del(ctx, key)

	_, err := r.SAdd(ctx, key, values...)

	assert.NoError(t, err)

	value, err := r.SRandMember(ctx, key)

	assert.NoError(t, err)
	assert.True(t, string(value) >= "1" && string(value) <= "3")
	members, err := r.SMembers(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, len(values), len(members))

	value, err = r.SPop(ctx, key)
	assert.NoError(t, err)
	assert.True(t, string(value) >= "1" && string(value) <= "3")
	time.Sleep(100 * time.Millisecond)
	members, err = r.SMembers(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, len(values)-1, len(members))
}

func TestSCard(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	key := "TestSCard"
	_, _ = r.Del(ctx, key)
	_, _ = r.SAdd(ctx, key, "value")

	//主从同步延迟
	time.Sleep(100 * time.Millisecond)
	if n, err := r.SCard(ctx, key); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestSPop(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	key := "TestSPop"
	_, _ = r.Del(ctx, key)
	_, _ = r.SAdd(ctx, key, "value")

	//主从同步延迟
	time.Sleep(100 * time.Millisecond)
	if item, err := r.SPop(ctx, key); err != nil {
		t.Error(err)
	} else if item == nil {
		t.Fail()
	} else if string(item) != "value" {
		t.Fail()
	}
	if item, _ := r.SPop(ctx, key); item != nil {
		t.Fail()
	}
}

func TestSRem(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	key := "TestSRem"
	_, _ = r.Del(ctx, key)
	_, _ = r.SAdd(ctx, key, "one", "two", "three")

	//主从同步延迟
	time.Sleep(100 * time.Millisecond)
	if n, err := r.SRem(ctx, key, "one", "four"); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Fail()
	}
}

func TestSScan(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	key := "TestSScan"
	r.Del(ctx, key)
	r.SAdd(ctx, key, "one", "two", "three")
	//主从同步延迟
	time.Sleep(100 * time.Millisecond)
	if _, list, err := r.SScan(ctx, key, 0, "", 0); err != nil {
		t.Error(err)
	} else if len(list) == 0 {
		t.Fail()
	}
}
