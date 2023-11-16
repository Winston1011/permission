package course

import (
	"permission/service"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
)

// 发送MQ消息示例
func BuyCourse(ctx *gin.Context) {
	var params struct {
		Username string `json:"username" binding:"omitempty"`
		UserID   string `json:"userID" binding:"required,max=20,min=1"`
		CourseID string `json:"courseID" binding:"required,max=50,min=1"`
		// 注意，当字段类型为string时,min/max指的是字符串的长度; 当字段类型为整型时,min/max指的是数值的大小
		Password string `json:"password" binding:"omitempty,numeric,min=10,max=32"`
	}
	// json 形式接收参数
	if err := ctx.ShouldBindJSON(&params); err != nil {
		zlog.Error(ctx, "[BuyCourse] params error: ", err.Error())
		// 用户未定义错误，默认输出错误码=-1，错误信息=err.Error()
		// base.RenderJsonFail(ctx, err)

		// 输出用户定义的特定错误码及错误信息，data为 {} 形式
		base.RenderJsonFail(ctx, err)
		return
	}

	// 购买课程核心逻辑
	c, err := service.BuyCourse(ctx, params.UserID, params.CourseID)
	if err != nil {
		base.RenderJsonFail(ctx, err)
		return
	}

	base.RenderJsonSucc(ctx, c)
}
