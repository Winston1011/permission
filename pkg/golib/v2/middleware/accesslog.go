package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const (
	_defaultPrintRequestLen  = 10240
	_defaultPrintResponseLen = 10240
)

type customRespWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w customRespWriter) WriteString(s string) (int, error) {
	if w.body != nil {
		w.body.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w customRespWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// access日志打印
type LoggerConfig struct {
	// SkipPaths is a url path array which logs are not written.
	SkipPaths []string `yaml:"skipPaths"`

	// request body 最大长度展示，0表示采用默认的10240，-1表示不打印
	MaxReqBodyLen int `yaml:"maxReqBodyLen"`
	// response body 最大长度展示，0表示采用默认的10240，-1表示不打印。指定长度的时候需注意，返回的json可能被截断
	MaxRespBodyLen int `yaml:"maxRespBodyLen"`

	// 自定义Skip功能
	Skip func(ctx *gin.Context) bool
}

func AccessLog(conf LoggerConfig) gin.HandlerFunc {
	notLogged := conf.SkipPaths
	var skip map[string]struct{}
	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	maxReqBodyLen := conf.MaxReqBodyLen
	if maxReqBodyLen == 0 {
		maxReqBodyLen = _defaultPrintRequestLen
	}

	maxRespBodyLen := conf.MaxRespBodyLen
	if maxRespBodyLen == 0 {
		maxRespBodyLen = _defaultPrintResponseLen
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path

		// body writer
		blw := &customRespWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 请求参数，涉及到回写，要在处理业务逻辑之前
		reqParam := getReqBody(c, maxReqBodyLen)

		c.Set(zlog.ContextKeyUri, path)
		_ = zlog.GetLogID(c)
		_ = zlog.GetRequestID(c)

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; ok {
			return
		}

		if conf.Skip != nil && conf.Skip(c) {
			return
		}

		// Stop timer
		end := time.Now()

		response := ""
		if blw.body != nil && maxRespBodyLen != -1 {
			response = blw.body.String()
			if len(response) > maxRespBodyLen {
				response = response[:maxRespBodyLen]
			}
		}

		// 固定notice
		commonFields := []zlog.Field{
			zlog.String("cuid", getReqValueByKey(c, "cuid")),
			zlog.String("device", getReqValueByKey(c, "device")),
			zlog.String("channel", getReqValueByKey(c, "channel")),
			zlog.String("os", getReqValueByKey(c, "os")),
			zlog.String("vc", getReqValueByKey(c, "vc")),
			zlog.String("vcname", getReqValueByKey(c, "vcname")),
			zlog.String("userid", getReqValueByKey(c, "userid")),
			zlog.String("host", c.Request.Host),
			zlog.String("method", c.Request.Method),
			zlog.String("httpProto", c.Request.Proto),
			zlog.String("handle", c.HandlerName()),
			zlog.String("userAgent", c.Request.UserAgent()),
			zlog.String("refer", c.Request.Referer()),
			zlog.String("clientIp", utils.GetClientIp(c)),
			zlog.String("cookie", getCookie(c)),
			zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
			zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
			zlog.Float64("cost", utils.GetRequestCost(start, end)),
			zlog.String("requestParam", reqParam),
			zlog.Int("responseStatus", c.Writer.Status()),
			zlog.String("response", response),
			zlog.Int("bodySize", c.Writer.Size()),
			zlog.String("reqModule", c.GetHeader("X-ZYB-REFERER-APP")), // 请求来源app
			zlog.String("reqErr", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		}

		// 新的notice添加方式
		customerFields := zlog.GetCustomerFields(c)
		// 老方式用户添加的k-v值
		for k, v := range zlog.GetCustomerKeyValue(c) {
			customerFields = append(customerFields, zlog.Reflect(k, v))
		}

		commonFields = append(commonFields, customerFields...)
		zlog.InfoLogger(c, "notice", commonFields...)
	}
}

// 请求参数
func getReqBody(c *gin.Context, maxReqBodyLen int) (reqBody string) {
	// 不打印参数
	if maxReqBodyLen == -1 {
		return reqBody
	}

	// body中的参数
	if c.Request.Body != nil && c.ContentType() == binding.MIMEMultipartPOSTForm {
		requestBody, err := c.GetRawData()
		if err != nil {
			zlog.WarnLogger(c, "get http request body error: "+err.Error())
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		if _, err := c.MultipartForm(); err != nil {
			zlog.WarnLogger(c, "parse http request form body error: "+err.Error())
		}
		reqBody = c.Request.PostForm.Encode()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	} else if c.Request.Body != nil && c.ContentType() == "application/octet-stream" {

	} else if c.Request.Body != nil {
		requestBody, err := c.GetRawData()
		if err != nil {
			zlog.WarnLogger(c, "get http request body error: "+err.Error())
		}
		reqBody = utils.BytesToString(requestBody)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	// 拼接上 url?rawQuery 的参数 todo 为了兼容以前逻辑，感觉参数应该分开写更好?
	if c.Request.URL.RawQuery != "" {
		reqBody += "&" + c.Request.URL.RawQuery
	}

	// 截断参数
	if len(reqBody) > maxReqBodyLen {
		reqBody = reqBody[:maxReqBodyLen]
	}

	return reqBody
}

// 从request body中解析特定字段作为notice key打印
func getReqValueByKey(ctx *gin.Context, k string) string {
	if vs, exist := ctx.Request.Form[k]; exist && len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func getCookie(ctx *gin.Context) string {
	cStr := ""
	for _, c := range ctx.Request.Cookies() {
		cStr += fmt.Sprintf("%s=%s&", c.Name, c.Value)
	}
	return strings.TrimRight(cStr, "&")
}

// access 添加kv打印
func AddNotice(k string, v interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		zlog.AddNotice(c, k, v)
		c.Next()
	}
}

func AddField(field ...zlog.Field) gin.HandlerFunc {
	return func(c *gin.Context) {
		zlog.AddField(c, field...)
		c.Next()
	}
}

func LoggerBeforeRun(ctx *gin.Context) {
	customCtx := ctx.CustomContext
	fields := []zlog.Field{
		zlog.String("handle", customCtx.HandlerName()),
		zlog.String("type", customCtx.Type),
	}

	zlog.InfoLogger(ctx, "start", fields...)
}

func LoggerAfterRun(ctx *gin.Context) {
	customCtx := ctx.CustomContext
	cost := utils.GetRequestCost(customCtx.StartTime, customCtx.EndTime)
	if customCtx.Error != nil {
		base.StackLogger(ctx, customCtx.Error)
	}

	// 用户自定义notice
	notices := zlog.GetCustomerKeyValue(ctx)

	var fields []zlog.Field
	for k, v := range notices {
		fields = append(fields, zlog.Reflect(k, v))
	}

	errMsg := "<nil>"
	if customCtx.Error != nil {
		errMsg = customCtx.Error.Error()
	}
	fields = append(fields,
		zlog.String("handle", customCtx.HandlerName()),
		zlog.String("type", customCtx.Type),
		zlog.Float64("cost", cost),
		zlog.String("desc", customCtx.Desc),
		zlog.String("error", errMsg),
	)

	zlog.InfoLogger(ctx, "end", fields...)
}
