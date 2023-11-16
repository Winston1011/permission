package user

/*
	分库分表示例
*/

import (
	"permission/components"
	"permission/helpers"
	"permission/models/demo"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID         int              `gorm:"primary_key;column:id" json:"id"`
	UserID     int              `gorm:"primary_key;column:user_id" json:"userID"`
	UserName   string           `gorm:"column:user_name" json:"userName"`
	UserPhone  demo.PhoneNumber `gorm:"column:user_phone" json:"userPhone"`
	Age        int              `gorm:"column:age" json:"age"`
	CreateTime demo.UnixTime    `gorm:"column:create_time" json:"createTime"`
}

/*
	假设 user 表拆为2张表：user1/user2
	分库分表规则:
	userID 为奇数存数到user1，userID为偶数存储到user2
*/

func toTableName(userID int) string {
	if userID%2 == 0 {
		return "user2"
	}
	return "user1"
}

func (u *User) Insert(ctx *gin.Context) (err error) {
	err = helpers.MysqlClientDemo.WithContext(ctx).Table(toTableName(u.UserID)).Create(u).Error
	if err != nil {
		return components.ErrorDbInsert.WrapPrintf(err, "demo insert error, input: %+v", u)
	}

	return nil
}

func GetUserByUserID(ctx *gin.Context, userID int) (list []User, err error) {
	err = helpers.MysqlClientDemo.Table(toTableName(userID)).Where("user_id = ?", userID).WithContext(ctx).Find(&list).Error
	if err != nil {
		return list, components.ErrorDbSelect.Wrap(err)
	}
	return list, nil
}

func GetUserByUserIDList(ctx *gin.Context, userIDList []int) (userList []User, err error) {
	tableMap := make(map[string][]int, 2)
	for _, u := range userIDList {
		tableMap[toTableName(u)] = append(tableMap[toTableName(u)], u)
	}

	db := helpers.MysqlClientDemo.WithContext(ctx)
	for t, u := range tableMap {
		var list []User
		err = db.Table(t).Where("user_id in ?", u).Find(&list).Error
		if err != nil {
			return nil, components.ErrorDbSelect.Wrap(err)
		}

		userList = append(userList, list...)
	}

	return userList, nil
}
