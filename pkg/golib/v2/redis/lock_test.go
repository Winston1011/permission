package redis

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedis_SetNxByEX(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	type args struct {
		key    string
		value  interface{}
		expire uint64
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "setLock",
			args: args{
				key:    "setex",
				value:  "1",
				expire: 20,
			},
		},
		{
			name: "setLocked",
			args: args{
				key:    "setex2",
				value:  "2",
				expire: 20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.SetNxByEX(ctx, tt.args.key, tt.args.value, tt.args.expire)
			assert.NoError(t, err)
			assert.True(t, res)
		})
	}
}

func TestSetNxByPX(t *testing.T) {
	setup()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	type args struct {
		key    string
		value  interface{}
		expire uint64
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "setLock",
			args: args{
				key:    "setex",
				value:  "1",
				expire: 20000,
			},
		},
		{
			name: "setLocked",
			args: args{
				key:    "setex",
				value:  "2",
				expire: 20000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := r.SetNxByPX(ctx, tt.args.key, tt.args.value, tt.args.expire)
			if err != nil {
				t.Errorf("   SetNxByPX() error = %v, res %v", err, res)
				return
			}
			if res == true {
				t.Errorf("   SetNxByPX() error = %v, res %v", err, res)
			} else {
				t.Errorf("   SetNxByPX() error = %v, res %v", err, res)
			}
		})
	}
}
