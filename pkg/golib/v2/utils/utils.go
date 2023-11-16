package utils

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Int64sContain(a []int64, x int64) bool {
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}

// 获取 uuid
func GenUUID() string {
	id := uuid.New()
	pass := hex.EncodeToString(id[:])
	tt := strconv.FormatInt(time.Now().Unix(), 10)
	return MultiJoinString("rpc", pass, tt)
}

// 获取函数名称
func GetFunctionName(i interface{}, seps ...rune) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	// 用 seps 进行分割
	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})

	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return ""
}

/*
  获取随机数
  不传参：0-100
  传1个参数：0-指定参数
  传2个参数：第1个参数-第2个参数
*/

func RandNum(num ...int) int {
	var start, end int
	if len(num) == 0 {
		start = 0
		end = 100
	} else if len(num) == 1 {
		start = 0
		end = num[0]
	} else {
		start = num[0]
		end = num[1]
	}

	rRandNumUtils := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rRandNumUtils.Intn(end-start+1) + start
}

func GetHandler(ctx *gin.Context) (handler string) {
	if ctx != nil {
		handler = ctx.HandlerName()
	}
	return handler
}

func JoinArgs(showByte int, args ...interface{}) string {
	var sumLen int

	argStr := make([]string, len(args))
	for i, v := range args {
		if s, ok := v.(string); ok {
			argStr[i] = s
		} else {
			argStr[i] = fmt.Sprintf("%v", v)
		}

		sumLen += len(argStr[i])
		if sumLen >= showByte {
			break
		}
	}

	argVal := strings.Join(argStr, " ")
	if sumLen > showByte {
		argVal = argVal[:showByte] + " ..."
	}
	return argVal
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
