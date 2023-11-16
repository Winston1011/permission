package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	h "permission/helpers"
	m "permission/models"
	"permission/service/node"
)

type GetListInput struct {
	AppId     int64
	ProductId int64
	UserId    int64
}

type GetListOutput struct {
	MenuList []m.Node `json:"menuList"`
	NodeList []m.Node `json:"nodeList"`
}

func (gi *GetListInput) GetMenuNodeList(ctx *gin.Context) (GetListOutput, error) {
	if err := gi.checkParams(); err != nil {
		return GetListOutput{}, err
	}
	var getListOutput GetListOutput
	nodeListInput := &node.NodeListInput{
		ProductId: gi.ProductId,
		AppId:     gi.AppId,
		UserId:    gi.UserId,
		NodeType:  components.NODE_TYPE_API,
	}
	retNodeList, _ := nodeListInput.GetNodeList(ctx)
	nodeListInput.NodeType = components.NODE_TYPE_PAGE
	retMenuList, _ := nodeListInput.GetNodeList(ctx)
	getListOutput.MenuList = retMenuList.NodeList
	getListOutput.NodeList = retNodeList.NodeList
	return getListOutput, nil
}

func (gi *GetListInput) checkParams() error {
	if gi.ProductId <= 0 {
		return h.NewError(components.ErrorGroupParamsInvalid, "productId 不合法")
	}
	if gi.AppId <= 0 {
		return h.NewError(components.ErrorGroupParamsInvalid, "appId 不合法")
	}
	if gi.UserId <= 0 {
		return h.NewError(components.ErrorGroupParamsInvalid, "userId 不合法")
	}
	return nil
}
