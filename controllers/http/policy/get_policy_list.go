package policy

import (
	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	p "permission/service/policy"
)

func GetPolicyList(ctx *gin.Context) {
	// 包含了条件查询 可以根据app id  product id 权限组进行筛选
	//var params struct {
	//	AppId     int64 `json:"appId" form:"appId"`
	//	ProductId int64 `json:"productId" form:"productId"`
	//	GroupId   int   `json:"groupId" form:"groupId"`
	//}
	//if err := ctx.BindQuery(&params); err != nil {
	//	zlog.Warnf(ctx, "policy params invalid err:%v", err)
	//	base.RenderJsonFail(ctx, components.ErrorPolicyParamsInvalid)
	//	return
	//}
	li := &p.ListInput{}
	response, err := li.GetPolicyList(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}

}
