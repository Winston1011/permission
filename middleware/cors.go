package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
	跨域中间件示例
	注意：使用gin在router中加载Cors()的时候，有两种方式：
	* 在所有分组路由的前面：比如，engine.Use(middleware.Cors)。这种情况下处理顺序为：cors -> 路由匹配, cors匹配到OPTIONS后返回204。
	该方法简单明了，对安全性要求不高或者本模块内几乎所有接口都需要考虑跨域问题可以使用该方法。
	* 在分组路由里：比如：group.Use(cors) 。 这种情况下处理顺序为：路由匹配 -> cors，router路由匹配会因找不到对应路径而返回404。
  	此时需要在group里实现OPTION方法：
	group.OPTIONS("/your-route", func(context *gin.Context) {
		context.AbortWithStatus(http.StatusNoContent)
		return
	})
   如果项目中只有个别接口需要考虑跨域问题，优先考虑使用功能该方法。
*/
func Cors(ctx *gin.Context) {
	origin := ctx.Request.Header.Get("Origin")
	if origin != "" {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, cache-control, X-CSRF-Token, Token,session")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Set("content-type", "application/json")
	}

	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}
	ctx.Next()
}
