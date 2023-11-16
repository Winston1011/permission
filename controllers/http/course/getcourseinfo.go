package course

import (
	"permission/api"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
)

// api 调用示例
func GetCourseInfo(ctx *gin.Context) {
	var params struct {
		CourseID string `form:"courseID" binding:"required,max=50,min=1"`
	}

	// get传参示例
	if err := ctx.ShouldBindQuery(&params); err != nil {
		zlog.Error(ctx, "[GetUserInfo] params error: ", err.Error())
		// 用户未定义错误，默认输出错误码=-1，错误信息=err.Error()
		base.RenderJsonFail(ctx, err)

		// 输出用户定义的特定错误码及错误信息，data为 {} 形式
		// base.RenderJsonFail(ctx, components.ErrorParamInvalid)
		return
	}

	// do something

	// 调用下游服务获取人员详细信息
	_, teacherInfo, err := api.GetUserList(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
		return
	}

	info := map[string]interface{}{
		"courseID":    params.CourseID,
		"courseName":  "math",
		"teacherInfo": teacherInfo,
	}

	base.RenderJsonSucc(ctx, info)
}
