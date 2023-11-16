package zlog

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/utils/metadata"
)

// util key
const (
	ContextKeyRequestID = "requestId"
	ContextKeyLogID     = "logID"
	ContextKeyNoLog     = "_no_log"
	ContextKeyUri       = "_uri"
	zapLoggerAddr       = "_zap_addr"
	sugaredLoggerAddr   = "_sugared_addr"
	customerFieldKey    = "__customerFields"
)

// header key
const (
	TraceHeaderKey      = "Uber-Trace-Id"
	LogIDHeaderKey      = "X_BD_LOGID"
	LogIDHeaderKeyLower = "x_bd_logid"
)

// 兼容虚拟机调用项目logid串联问题
func GetLogID(ctx *gin.Context) string {
	if ctx == nil {
		return genLogID()
	}

	// 上次获取到的
	if logID := ctx.GetString(ContextKeyLogID); logID != "" {
		return logID
	}

	// 尝试从header中获取
	var logID string
	if ctx.Request != nil && ctx.Request.Header != nil {
		logID = ctx.GetHeader(LogIDHeaderKey)
		if logID == "" {
			logID = ctx.GetHeader(LogIDHeaderKeyLower)
		}
	}

	if logID == "" {
		logID = genLogID()
	}

	ctx.Set(ContextKeyLogID, logID)
	return logID
}

func GetRequestID(ctx *gin.Context) string {
	if ctx == nil {
		return genRequestID()
	}

	// 从ctx中获取
	if r := ctx.GetString(ContextKeyRequestID); r != "" {
		return r
	}

	// 优先从header中获取
	var requestID string
	if ctx.Request != nil && ctx.Request.Header != nil {
		requestID = ctx.Request.Header.Get(TraceHeaderKey)
	}

	// 新生成
	if requestID == "" {
		requestID = genRequestID()
	}

	ctx.Set(ContextKeyRequestID, requestID)
	return requestID
}

func genLogID() (requestId string) {
	// 随机生成
	usec := uint64(time.Now().UnixNano())
	requestId = strconv.FormatUint(usec&0x7FFFFFFF|0x80000000, 10)
	return requestId
}

var generator = utils.NewRand(time.Now().UnixNano())

func genRequestID() string {
	// 生成 uint64的随机数, 并转换成16进制表示方式
	number := uint64(generator.Int63())
	traceID := fmt.Sprintf("%016x", number)

	var buffer bytes.Buffer
	buffer.WriteString(traceID)
	buffer.WriteString(":")
	buffer.WriteString(traceID)
	buffer.WriteString(":0:1")
	return buffer.String()
}

func GetRequestUri(ctx *gin.Context) string {
	if ctx == nil {
		return ""
	}
	return ctx.GetString(ContextKeyUri)
}

// 用户自定义Notice
func AddNotice(ctx *gin.Context, key string, val interface{}) {
	if meta, ok := metadata.CtxFromGinContext(ctx); ok {
		if n := metadata.Value(meta, metadata.Notice); n != nil {
			if _, ok = n.(map[string]interface{}); ok {
				notices := n.(map[string]interface{})
				notices[key] = val
			}
		}
	}
}

// a new method for customer notice
func AddField(c *gin.Context, field ...Field) {
	customerFields := GetCustomerFields(c)
	if customerFields == nil {
		customerFields = field
	} else {
		customerFields = append(customerFields, field...)
	}

	c.Set(customerFieldKey, customerFields)
}

// 获得所有用户自定义的Field
func GetCustomerFields(c *gin.Context) (customerFields []Field) {
	if v, exist := c.Get(customerFieldKey); exist {
		customerFields, _ = v.([]Field)
	}
	return customerFields
}

// 获得所有用户自定义的Notice
func GetCustomerKeyValue(ctx *gin.Context) map[string]interface{} {
	meta, ok := metadata.CtxFromGinContext(ctx)
	if !ok {
		return nil
	}

	n := metadata.Value(meta, metadata.Notice)
	if n == nil {
		return nil
	}
	if notices, ok := n.(map[string]interface{}); ok {
		return notices
	}

	return nil
}

// server.log 中打印出用户自定义Notice
func PrintNotice(ctx *gin.Context) {
	notices := GetCustomerKeyValue(ctx)

	var fields []interface{}
	for k, v := range notices {
		fields = append(fields, k, v)
	}
	sugaredLogger(ctx).With(fields...).Info("notice")
}

// server.log 中打印出用户添加的所有Field
func PrintFields(ctx *gin.Context) {
	fields := GetCustomerFields(ctx)
	zapLogger(ctx).Info("notice", fields...)
}

func SetNoLogFlag(ctx *gin.Context) {
	ctx.Set(ContextKeyNoLog, true)
}

func SetLogFlag(ctx *gin.Context) {
	ctx.Set(ContextKeyNoLog, false)
}

func noLog(ctx *gin.Context) bool {
	if ctx == nil {
		return false
	}
	flag, ok := ctx.Get(ContextKeyNoLog)
	if ok && flag == true {
		return true
	}
	return false
}
