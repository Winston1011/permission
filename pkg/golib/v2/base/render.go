package base

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"permission/pkg/golib/v2/zlog"
)

// 定义render通用类型
type Render interface {
	SetReturnCode(int)
	SetReturnMsg(string)
	SetReturnData(interface{})
	GetReturnCode() int
	GetReturnMsg() string
}

var newRender func() Render

func RegisterRender(s func() Render) {
	newRender = s
}

func newJsonRender() Render {
	if newRender == nil {
		newRender = defaultNew
	}
	return newRender()
}

func RenderJson(ctx *gin.Context, code int, msg string, data interface{}) {
	r := newJsonRender()
	r.SetReturnCode(code)
	r.SetReturnMsg(msg)
	r.SetReturnData(data)

	setCommonHeader(ctx, code, msg)
	ctx.JSON(http.StatusOK, r)
	return
}

func RenderJsonSucc(ctx *gin.Context, data interface{}) {
	r := newJsonRender()
	r.SetReturnCode(0)
	r.SetReturnMsg("succ")
	r.SetReturnData(data)

	setCommonHeader(ctx, 0, "succ")
	ctx.JSON(http.StatusOK, r)
	return
}

func RenderJsonFail(ctx *gin.Context, err error) {
	r := newJsonRender()

	code, msg := -1, errors.Cause(err).Error()
	switch errors.Cause(err).(type) {
	case Error:
		code = errors.Cause(err).(Error).ErrNo
		msg = errors.Cause(err).(Error).ErrMsg
	default:
	}

	r.SetReturnCode(code)
	r.SetReturnMsg(msg)
	r.SetReturnData(gin.H{})

	setCommonHeader(ctx, code, msg)
	ctx.JSON(http.StatusOK, r)

	// 打印错误栈
	StackLogger(ctx, err)
	return
}

func RenderJsonAbort(ctx *gin.Context, err error) {
	r := newJsonRender()

	switch errors.Cause(err).(type) {
	case Error:
		r.SetReturnCode(errors.Cause(err).(Error).ErrNo)
		r.SetReturnMsg(errors.Cause(err).(Error).ErrMsg)
		r.SetReturnData(gin.H{})
	default:
		r.SetReturnCode(-1)
		r.SetReturnMsg(errors.Cause(err).Error())
		r.SetReturnData(gin.H{})
	}

	setCommonHeader(ctx, r.GetReturnCode(), r.GetReturnMsg())
	ctx.AbortWithStatusJSON(http.StatusOK, r)

	return
}

// default render

var defaultNew = func() Render {
	return &DefaultRender{}
}

type DefaultRender struct {
	ErrNo  int         `json:"errNo"`
	ErrMsg string      `json:"errMsg"`
	Data   interface{} `json:"data"`
}

func (r *DefaultRender) GetReturnCode() int {
	return r.ErrNo
}
func (r *DefaultRender) SetReturnCode(code int) {
	r.ErrNo = code
}
func (r *DefaultRender) GetReturnMsg() string {
	return r.ErrMsg
}
func (r *DefaultRender) SetReturnMsg(msg string) {
	r.ErrMsg = msg
}
func (r *DefaultRender) GetReturnData() interface{} {
	return r.Data
}
func (r *DefaultRender) SetReturnData(data interface{}) {
	r.Data = data
}

// 设置通用header头
func setCommonHeader(ctx *gin.Context, code int, _ string) {
	ctx.Header("X_BD_UPS_ERR_NO", strconv.Itoa(code))
	ctx.Header("Request-Id", zlog.GetRequestID(ctx))
}

// 打印错误栈
func StackLogger(ctx *gin.Context, err error) {
	ss := fmt.Sprintf("%+v", err)
	if !strings.Contains(ss, "\n") {
		return
	}

	stacks := strings.Split(ss, "\n")
	zlog.ErrorLogger(ctx, "errorstack", zlog.Strings("stack", stacks))
}
