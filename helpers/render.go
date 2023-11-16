package helpers

import (
	"permission/pkg/golib/v2/base"
)

func initRender() {
	base.RegisterRender(func() base.Render {
		return &CustomRender{}
	})
}

// 业务自定义render （一般只需修改json tag即可）
type CustomRender struct {
	ErrNo  int         `json:"errNo"`
	ErrMsg string      `json:"errStr"`
	Data   interface{} `json:"data"`
}

func (r *CustomRender) SetReturnCode(code int) {
	r.ErrNo = code
}
func (r *CustomRender) SetReturnMsg(msg string) {
	r.ErrMsg = msg
}
func (r *CustomRender) SetReturnData(data interface{}) {
	r.Data = data
}

func (r *CustomRender) GetReturnCode() int {
	return r.ErrNo
}
func (r *CustomRender) GetReturnMsg() string {
	return r.ErrMsg
}
func (r *CustomRender) GetReturnData() interface{} {
	return r.Data
}
