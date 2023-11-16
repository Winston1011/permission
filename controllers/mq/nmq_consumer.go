package mq

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/zlog"
)

// nmq 消费者测试示例
func NmqTest(ctx *gin.Context) {
	// transid, cmd, topic等信息都在msg中
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.String(http.StatusBadRequest, "read mq body fail")
		return
	}

	nmqMsg, _ := mcpack.Decode(body) // 解码mapack消息体

	// nmqMsg为map[string]interface, 后续的处理逻辑保持与之前一致即可
	zlog.Debugf(ctx, "got nmq message: %+v", nmqMsg)

	ctx.String(http.StatusOK, "success")
}
