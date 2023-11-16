package node

import (
	"git.zuoyebang.cc/pkg/golib/v2/zlog"
	"github.com/gin-gonic/gin"
	"permission/components"
	h "permission/helpers"
	m "permission/models"
)

type NodeListInput struct {
	AppId     int64
	ProductId int64
	UserId    int64
	GroupId   int64
	NodeType  int8
}

type ListOutput struct {
	NodeList []m.Node `json:"nodeList"`
}

func (li *NodeListInput) GetNodeList(ctx *gin.Context) (ret ListOutput, err error) {
	if err = li.checkParams(); err != nil {
		return ret, err
	}
	err, nodeTree := li.getNodeTreeMap(ctx)
	if err != nil {
		zlog.Warnf(ctx, "get nodeTree failure")
		return ret, h.NewError(components.ErrorDbSelect, "get nodeTree failure")
	}
	nodes := nodeTree[0]
	for i := 0; i < len(nodes); i++ {
		li.getChildrenList(&nodes[i], nodeTree)
	}
	// 处理选中node
	_, checkedNode := li.getCheckedNodes(ctx)
	for i := 0; i < len(nodes); i++ {
		updateCheckedStatus(&nodes[i], checkedNode)
	}
	ret.NodeList = nodes
	return ret, err
}

func updateCheckedStatus(node *m.Node, groupNodes []m.GroupNode) {
	for _, v := range groupNodes {
		if node.ID == v.NodeId {
			node.Selected = 1
		}
	}
	children := node.Children
	if children != nil {
		for i := 0; i < len(children); i++ {
			updateCheckedStatus(&children[i], groupNodes)
		}
	}
}

func (li *NodeListInput) getNodeTreeMap(ctx *gin.Context) (err error, treeMap map[int64][]m.Node) {
	var node m.Node
	// 获取所有node
	condition := map[string]interface{}{
		"node_type": li.NodeType,
	}
	nodeList, err := node.GetNodeListByCondition(ctx, condition)
	treeMap = make(map[int64][]m.Node)
	for _, v := range nodeList {
		treeMap[v.ParentID] = append(treeMap[v.ParentID], v)
	}
	return err, treeMap
}

func (li *NodeListInput) getChildrenList(node *m.Node, treeMap map[int64][]m.Node) (err error) {
	node.Children = treeMap[node.ID]
	for i := 0; i < len(node.Children); i++ {
		err = li.getChildrenList(&node.Children[i], treeMap)
	}
	return err
}

func (li *NodeListInput) checkParams() error {
	if li.AppId < 0 {
		return h.NewError(components.ErrorNodeParamsInvalid, "appId 不合法")
	}
	if li.ProductId < 0 {
		return h.NewError(components.ErrorNodeParamsInvalid, "productId 不合法")
	}
	return nil
}

func (li *NodeListInput) getCheckedNodes(ctx *gin.Context) (err error, nodes []m.GroupNode) {
	// 获取已选node
	if li.UserId > 0 && li.GroupId == 0 {
		// userId -> groupId
		userGroup := &m.UserGroup{
			UserId: li.UserId,
		}
		condition := map[string]interface{}{
			"product_id": li.ProductId,
			"app_id":     li.AppId,
			"user_id":    li.UserId,
			"user_type":  components.USER_TYPE_OUTER,
		}
		userGroupInfo, _ := userGroup.GetUserGroupByCondition(ctx, condition)
		li.GroupId = userGroupInfo.GroupId
	}
	groupNode := m.GroupNode{
		GroupId: li.GroupId,
	}
	if li.GroupId > 0 {
		condition := map[string]interface{}{
			"group_id":  li.GroupId,
			"node_type": li.NodeType,
		}
		nodes, err = groupNode.GetGroupNodeListByConds(ctx, condition)
	}
	return err, nodes
}
