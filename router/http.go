package router

import (
	"github.com/gin-gonic/gin"
	"permission/controllers/http/group"
	"permission/controllers/http/node"
	"permission/controllers/http/perm"
	"permission/controllers/http/policy"
	"permission/controllers/http/user"
	"permission/middleware"
	m "permission/pkg/golib/v2/middleware"
)

func Http(engine *gin.Engine) {

	router := engine.Group("/permission")
	//
	// 通用中间件
	router.Use(m.AddNotice("globalCustomerNotice", "permission"))

	router.Use(middleware.Recover)

	// console服务请求资源鉴权
	checkGroup := router.Group("/request", m.AddNotice("customerNotice", "v1"))
	{
		checkGroup.POST("/checkpermission", perm.CheckPermission)
	}

	// 权限组设置
	permGroup := router.Group("group", m.AddNotice("customerNotice", "v1"))
	{
		permGroup.POST("/creategroup", group.CreateGroup)
		permGroup.POST("/deletegroup", group.DeleteGroup)
		permGroup.POST("/updategroup", group.Updategroup)
		permGroup.POST("/getgrouplist", group.GetGroupList)
		permGroup.POST("/getmenunodelist", group.GetMenuNodeList)
	}

	// 校验规则管理
	policyManager := router.Group("policy", m.AddNotice("customerNotice", "v1"))
	{
		policyManager.POST("/createpolicy", policy.CreatePolicy)
		policyManager.POST("/stoppolicy", policy.StopPolicy)
		policyManager.POST("/deletepolicy", policy.DeletePolicy)
		policyManager.POST("/getpolicylist", policy.GetPolicyList)
	}

	// 路由页面、接口管理
	nodeGroup := router.Group("node", m.AddNotice("customerNotice", "v1"))
	{
		nodeGroup.POST("/createnode", node.CreateNode)
		nodeGroup.POST("/updatenode", node.UpdateNode)
		nodeGroup.POST("/deletenode", node.DeleteNode)
		nodeGroup.POST("/getnodelist", node.GetNodeList)
	}

	// 用户权限组设置
	userPermGroup := router.Group("user", m.AddNotice("customerNotice", "v1"))
	{
		userPermGroup.POST("/addrelusergroup", user.CreateRelUserGroup)
	}
}
