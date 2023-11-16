package group

import (
	"github.com/gin-gonic/gin"
	"permission/components"
	m "permission/models"
)

// GetInput 加条件查询
type GetInput struct {
}

type ListOutput struct {
	GroupList []m.Group `json:"groupList"`
}

func (gi *GetInput) GetGroupsList(ctx *gin.Context) (ret ListOutput, err error) {
	err, groupTree := gi.getGroupTreeMap(ctx)
	groups := groupTree[0]
	for i := 0; i < len(groups); i++ {
		gi.getChildrenList(&groups[i], groupTree)
	}
	ret.GroupList = groups
	return ret, err
}

func (gi *GetInput) getGroupTreeMap(ctx *gin.Context) (err error, treeMap map[int64][]m.Group) {
	var group m.Group
	condition := map[string]interface{}{
		"status": components.GROUP_STATUS_ACTIVE,
	}
	groupList, err := group.GetGroupListByConds(ctx, condition)
	treeMap = make(map[int64][]m.Group)
	for _, v := range groupList {
		treeMap[v.ParentId] = append(treeMap[v.ParentId], v)
	}
	return err, treeMap
}

func (gi *GetInput) getChildrenList(group *m.Group, treeMap map[int64][]m.Group) (err error) {
	group.Children = treeMap[group.ID]
	for i := 0; i < len(group.Children); i++ {
		err = gi.getChildrenList(&group.Children[i], treeMap)
	}
	return err
}
