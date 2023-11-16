package router

import (
	"permission/controllers/http/course"
	"permission/controllers/http/user"
	"permission/middleware"

	"github.com/gin-gonic/gin"
	m "permission/pkg/golib/v2/middleware"
	"permission/pkg/golib/v2/zlog"
)

func Http(engine *gin.Engine) {
	router := engine.Group("/permission")

	// 通用中间件
	router.Use(m.AddField(zlog.String("globalCustomerNotice", "permission")))

	// per group middleware! in this case we use
	// m.AddField() middleware just in the "courseGroup" group.
	courseGroup := router.Group("/api/course", m.AddField(zlog.String("customerNotice", "v1")))
	{
		courseGroup.POST("/buy", course.BuyCourse)
		courseGroup.GET("/getinfo", course.GetCourseInfo)
	}

	// router group
	userGroup := router.Group("/api/user", middleware.AppCheck)
	{
		userGroup.POST("/getlist", user.GetUserInfoList)
		userGroup.POST("/get-user-course", user.GetUserCourse)
	}
}
