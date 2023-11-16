package group

import (
	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	g "permission/service/group"
)

func GetGroupList(ctx *gin.Context) {
	gi := &g.GetInput{}
	response, err := gi.GetGroupsList(ctx)
	if err != nil {
		base.RenderJsonFail(ctx, err)
	} else {
		base.RenderJsonSucc(ctx, response)
	}
}
