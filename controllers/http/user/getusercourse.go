package user

import (
	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/gomcpack/mcpack"
)

// 不建议使用mcPack打包，重构的时候最好去掉私有协议，推送下游服务一起改造
// post pack data 示例
func GetUserCourse(ctx *gin.Context) {
	var params struct {
		UserID   string `form:"userID" binding:"required,numeric,max=50,min=1"`
		UserName string `form:"userName" binding:"required,numeric,max=50,min=1"`
	}

	data, err := ctx.GetRawData()
	if err != nil {
		base.RenderJsonFail(ctx, err)
		return
	}
	if err := mcpack.Unmarshal(data, &params); err != nil {
		base.RenderJsonFail(ctx, err)
		return
	}

	// do something

	info := map[string]interface{}{
		"userID":   params.UserID,
		"username": params.UserName,
		"courseInfo": map[string]string{
			"courseID":   "6379",
			"courseName": "math",
		},
	}

	base.RenderJsonSucc(ctx, info)
}
