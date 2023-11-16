package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HttpXBDCallerURI         = "X_BD_CALLER_URI"
	HttpXBDCallerURIV2       = "HTTP_X_BD_CALLER_URI"
	HttpUrlPressureCallerKey = "_caller_uri"
	HttpUrlPressureMarkKey   = "_press_mark"
	ZYBTransportHeader       = "X-Zyb-Ctx-"
	NavigatorOCSURI          = "/qa/test"
)

// deprecated: use  IsNavigatorPressure or IsQATestPressure instead
func GetPressureFlag(ctx *gin.Context) (callerURI string, pressMark int) {
	callerURI = GetCallerURI(ctx)
	pressMark = 0
	if callerURI != "" && strings.Contains(callerURI, NavigatorOCSURI) {
		pressMark = 1
	}

	return callerURI, pressMark
}

// 是否是 全链路压测流量
func IsNavigatorPressure(ctx *gin.Context) bool {
	callerURI := GetCallerURI(ctx)
	arr := strings.Split(callerURI, ",")
	if len(arr) >= 3 {
		if arr[1] == "/qa/test" && arr[2] == "1" {
			return true
		}
	}
	return false
}

// 是否是 QA测试流量
func IsQATestPressure(ctx *gin.Context) bool {
	callerURI := GetCallerURI(ctx)
	arr := strings.Split(callerURI, ",")
	if len(arr) >= 2 {
		if arr[1] == NavigatorOCSURI {
			return true
		}
	}
	return false
}

// 获取压测时间 (注意，这里获取不到就返回0)
func GetPressureTime(ctx *gin.Context) (pressureTime int) {
	callerURI := GetCallerURI(ctx)
	arr := strings.Split(callerURI, ",")
	if len(arr) >= 7 {
		pressureTime, _ = strconv.Atoi(arr[6])
	}
	return pressureTime
}

// 获取callerURI
func GetCallerURI(ctx *gin.Context) (callerURI string) {
	if ctx == nil {
		return callerURI
	}
	if ctx.Request != nil {
		// 优先取header
		callerURI = ctx.GetHeader(HttpXBDCallerURI)
		if callerURI == "" {
			callerURI = ctx.GetHeader(HttpXBDCallerURIV2)
		}

		// 然后取query参数（nmq）
		if callerURI == "" {
			callerURI = ctx.Query(HttpUrlPressureCallerKey)
		}
	}

	// 然后取ctx中的k-v （nmq回调）
	if callerURI == "" {
		callerURI = ctx.GetString(HttpUrlPressureCallerKey)
	}
	return callerURI
}

// 获取所有需要透传的header
func GetTransportHeader(ctx *gin.Context) (transKey map[string]string) {
	if ctx == nil {
		return nil
	}

	if ctx.Request != nil && ctx.Request.Header != nil {
		transKey = make(map[string]string)
		for k, v := range ctx.Request.Header {
			if len(k) > 10 && k[0:10] == ZYBTransportHeader && len(v) > 0 {
				transKey[k] = v[0]
			}
		}
		return transKey
	}

	// 然后取ctx中的k-v （mq回调）
	return ctx.GetStringMapString(ZYBTransportHeader)
}

func SetTransportHeader(ctx *gin.Context, header map[string]string) {
	ctx.Set(ZYBTransportHeader, header)
}
