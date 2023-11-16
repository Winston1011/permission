package api

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
)

func decodeResponse(ctx *gin.Context, res *base.ApiResult, output interface{}) (errno int, err error) {
	var r base.DefaultRender
	if err = json.Unmarshal(res.Response, &r); err != nil {
		zlog.Errorf(res.Ctx, "http response decode err, err: %s", res.Response)
		return errno, err
	}

	// 注意这只是一个示例，这里做了业务errno的判断。如果不需要可以删除相关逻辑。
	errno = r.ErrNo
	if errno != 0 {
		zlog.Errorf(res.Ctx, "http response code: %d", r.ErrNo)
		return errno, err
	}

	/*
		一般接口输出应该是规范的，比如：resp := `{"errNo":0,"errMsg":"succ","data":{}}`
		对于早期的php返回不规范，可能会出现：resp := `{"errNo":0,"errMsg":"succ","data":[]}`
		也就是同一个接口的 data 返回两种类型。
		对于不规范的形式，使用 `mapstructure.Decode` 到struct会panic，因为类型不匹配。
		如果确定`{}` 是正常返回，可以增加以下判断来避免`[]`解析panic的问题。
		不需要删除即可
	*/
	if _, ok := r.Data.(map[string]interface{}); !ok {
		return errno, nil
	}

	if err := mapstructure.Decode(r.Data, &output); err != nil {
		zlog.Warnf(ctx, "api call data decode error: %s", err.Error())
		return errno, err
	}

	return errno, nil
}
