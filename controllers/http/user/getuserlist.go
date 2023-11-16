package user

import (
	"strconv"
	"strings"

	"permission/components"
	"permission/service"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
)

func GetUserInfoList(ctx *gin.Context) {
	// 之后的所有日志不打印了
	// zlog.SetNoLogFlag(ctx)

	var params struct {
		UserIDs    string `form:"userIDList" binding:"required"`
		UserIDList []int
	}

	// post 表单形式接收参数
	if err := ctx.ShouldBind(&params); err != nil {
		// 输出用户定义的特定错误码及错误信息，data为 {} 形式
		base.RenderJsonFail(ctx, components.ErrorParamInvalid)
		return
	}

	// showHeader(ctx)

	// 参数处理
	us := strings.Split(params.UserIDs, ",")
	for _, u := range us {
		i, err := strconv.Atoi(u)
		if err != nil {
			// 根据业务逻辑查看是返回还是忽略
			continue
		}
		params.UserIDList = append(params.UserIDList, i)
	}
	if len(params.UserIDList) == 0 {
		base.RenderJsonSucc(ctx, gin.H{"count": 0, "list": nil})
		return
	}

	// 查询用户信息
	list, err := service.GetUserInfo(ctx, params.UserIDList)
	if err != nil {
		base.RenderJsonFail(ctx, err)
		return
	}

	base.RenderJsonSucc(ctx, gin.H{
		"count": len(list),
		"list":  list,
	})
}

func showHeader(ctx *gin.Context) {
	// header 获取示例
	for k, v := range ctx.Request.Header {
		if len(v) > 0 {
			zlog.Debug(ctx, "header key=", k, " value=", v[0])
		}
	}

	// cookie 获取示例 (也可以直接从header中取 Cookie 字段)
	if ck, err := ctx.Request.Cookie("uname"); err == nil && ck != nil {
		zlog.Debug(ctx, "cookie name=", ck.Name, "value=", ck.Value)
	}

	for _, ck := range ctx.Request.Cookies() {
		if ck != nil {
			zlog.Debug(ctx, "cookie name=", ck.Name, "value=", ck.Value)
		}
	}
}
