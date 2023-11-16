package group

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/helpers"
	m "permission/models"
	"permission/pkg/golib/v2/zlog"
	"time"
)

type GUpdateInput struct {
	ProductId   int64
	AppId       int64
	GroupId     int64
	UserId      int64
	GroupName   string
	GroupStatus int8
	NodeList    []int64
	MenuList    []int64
}

func (gu *GUpdateInput) UpdateGroup(ctx *gin.Context) (bool, error) {
	if err := gu.checkParams(); err != nil {
		return false, err
	}
	group := &m.Group{
		ID:        gu.GroupId,
		GroupName: gu.GroupName,
		Status:    gu.GroupStatus,
		UpdateUid: gu.UserId,
	}
	//groupNode
	groupNode := &m.GroupNode{
		GroupId: gu.GroupId,
	}
	condition := map[string]interface{}{
		"group_id":  gu.GroupId,
		"node_type": components.NODE_TYPE_API,
	}
	groupNodeList, err := groupNode.GetGroupNodeListByConds(ctx, condition)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbSelect, "get groupNodeListByConds failure")
	}
	condition = map[string]interface{}{
		"group_id":  gu.GroupId,
		"node_type": components.NODE_TYPE_PAGE,
	}
	groupMenuList, err := groupNode.GetGroupNodeListByConds(ctx, condition)
	if err != nil {
		return false, helpers.NewError(components.ErrorDbSelect, "get groupMenuListByConds failure")
	}
	oldNodeList := ConvertId2Slice(groupNodeList)
	oldMenuList := ConvertId2Slice(groupMenuList)
	insertNodeIdList, deleteNodeIdList := gu.filtrateId(oldNodeList, gu.NodeList)
	insertMenuIdList, deleteMenuIdList := gu.filtrateId(oldMenuList, gu.MenuList)
	return gu.update(ctx, group, groupNode, insertNodeIdList, deleteNodeIdList, insertMenuIdList, deleteMenuIdList)
}

func (gu *GUpdateInput) update(ctx *gin.Context, group *m.Group, groupNode *m.GroupNode, insertNodeIdList []int64, deleteNodeIdList []int64, insertMenuIdList []int64, deleteMenuIdList []int64) (bool, error) {
	result, err := func() (bool, error) {
		node := &m.Node{}
		var txFlowErr error
		// 开始事务
		var tx = helpers.MysqlClientPermission.Begin()
		if err := tx.Error; err != nil {
			zlog.Warnf(ctx, "DB错误 开启事务失败", err)
			return false, helpers.NewError(components.ErrorDbError, err.Error())
		}
		defer func() {
			if txFlowErr != nil {
				// 回滚事务
				if _err := tx.Rollback().Error; _err != nil {
					zlog.Warnf(ctx, "DB错误 事务回滚失败", _err)
					txFlowErr = _err
					return
				}
			} else {
				// 提交事务
				if _err := tx.Commit().Error; _err != nil {
					zlog.Warnf(ctx, "DB错误 事务提交失败", _err)
					txFlowErr = _err
					return
				}
				helpers.Enforcer.LoadPolicy() //加载新的校验规则
			}
		}()
		// 1.更新group的信息
		updatesFields := map[string]interface{}{
			"group_name":  group.GroupName,
			"status":      group.Status,
			"update_uid":  group.UpdateUid,
			"update_time": time.Now().Unix(),
		}
		if _, err := group.UpdateGroupById(ctx, group.ID, updatesFields, tx); err != nil {
			txFlowErr = err
			zlog.Errorf(ctx, "update group fail, err:%v", err)
			return false, err
		}
		// 2.批量插入新的nodeId映射关系
		var insertNodeList []m.GroupNode
		var insertCasbinRules []m.CasbinRule
		if len(insertNodeIdList) > 0 {
			for _, v := range insertNodeIdList {
				insertNodeList = append(insertNodeList, m.GroupNode{
					GroupId:  gu.GroupId,
					NodeId:   v,
					NodeType: components.NODE_TYPE_API,
				})
				nodeInfo, err := node.GetNodeById(ctx, v)
				if err != nil {
					txFlowErr = err
					zlog.Errorf(ctx, "get node detail fail, err:%v", err)
					return false, err
				}
				insertCasbinRules = append(insertCasbinRules, m.CasbinRule{
					Ptype:           components.CASBIN_RULE_PTYPE,
					GroupId:         fmt.Sprintf("%d", gu.GroupId),
					ProductAppField: fmt.Sprintf("%d:%d", gu.ProductId, gu.AppId),
					Resource:        nodeInfo.Resource,
					PermissionType:  "any",
					Status:          "allow",
				})
			}
			if _, err := groupNode.BatchInsertGroupNode(ctx, insertNodeList, tx); err != nil {
				txFlowErr = err
				zlog.Errorf(ctx, "batch create rel group node fail, err:%v", err)
				return false, err
			}
			// 4.批量插入新的组员校验规则
			casbinRule := &m.CasbinRule{}
			if _, err := casbinRule.BatchUpsertCasbinRule(ctx, insertCasbinRules, tx); err != nil {
				txFlowErr = err
				zlog.Errorf(ctx, "batch create casbin rule fail, err:%v", err)
				return false, err
			}
		}
		if len(deleteNodeIdList) > 0 {
			// 3.批量删除被删除的nodeId映射关系
			if _, err := groupNode.BatchDeleteGroupNode(ctx, deleteNodeIdList, tx); err != nil {
				txFlowErr = err
				zlog.Errorf(ctx, "batch delete rel group node fail, err:%v", err)
				return false, err
			}
			// 5.批量删除被删除的资源权限校验规则
			for _, v := range deleteNodeIdList {
				nodeInfo, err := node.GetNodeById(ctx, v)
				if err != nil {
					txFlowErr = err
					zlog.Errorf(ctx, "get node detail fail, err:%v", err)
					return false, err
				}
				deleteModel := &m.CasbinRule{}
				condition := map[string]interface{}{
					"ptype": "p",
					"v0":    fmt.Sprintf("%d", gu.GroupId),
					"v1":    fmt.Sprintf("%d:%d", gu.ProductId, gu.AppId),
					"v2":    nodeInfo.Resource,
				}
				if _, err := deleteModel.DeleteCasbinRuleByCondition(ctx, condition, tx); err != nil {
					txFlowErr = err
					zlog.Errorf(ctx, "batch delete casbin rule fail, err:%v", err)
					return false, err
				}
			}
		}
		if len(insertMenuIdList) > 0 {
			var insertMenuList []m.GroupNode
			for _, v := range insertMenuIdList {
				insertMenuList = append(insertMenuList, m.GroupNode{
					GroupId:  gu.GroupId,
					NodeId:   v,
					NodeType: components.NODE_TYPE_PAGE,
				})
			}
			if _, err := groupNode.BatchInsertGroupNode(ctx, insertMenuList, tx); err != nil {
				txFlowErr = err
				zlog.Errorf(ctx, "batch create rel group menu fail, err:%v", err)
				return false, err
			}
		}
		if len(deleteMenuIdList) > 0 {
			// 3.批量删除被删除的menuId映射关系
			if _, err := groupNode.BatchDeleteGroupNode(ctx, deleteMenuIdList, tx); err != nil {
				txFlowErr = err
				zlog.Errorf(ctx, "batch delete rel group menu fail, err:%v", err)
				return false, err
			}
		}
		return true, txFlowErr
	}()
	return result, err
}

func ConvertId2Slice(groupNodeList []m.GroupNode) (oldNodeList []int64) {
	for _, v := range groupNodeList {
		oldNodeList = append(oldNodeList, v.NodeId)
	}
	return
}

func (gu *GUpdateInput) filtrateId(oldIdList []int64, newIdList []int64) (insertIdList []int64, deleteIdList []int64) {
	deleteIdList = helpers.Subtraction(oldIdList, newIdList)
	insertIdList = helpers.Subtraction(newIdList, oldIdList)
	return
}

func (gu *GUpdateInput) checkParams() error {
	if len(gu.GroupName) <= 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "groupName 不合法")
	}
	if gu.GroupId < 0 {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "groupId 不合法")
	}
	if gu.GroupStatus != components.GROUP_STATUS_ACTIVE && gu.GroupStatus != components.GROUP_STATUS_CLOSE {
		return helpers.NewError(components.ErrorGroupParamsInvalid, "groupStatus 不合法")
	}
	return nil
}
