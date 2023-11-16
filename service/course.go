/*
	课程 course 相关业务逻辑
	同一文件中的函数应按接收者分组，函数应按粗略的调用顺序排序。
	如果service中的业务逻辑复杂，建议把数据组装相关的逻辑放到data层处理，使得service层逻辑更简洁、清晰。
*/
package service

import (
	"fmt"
	"math/rand"

	"permission/models/user"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"permission/pkg/golib/v2/rmq"
	"permission/pkg/golib/v2/zlog"
)

type BuyCourseInfo struct {
	// 使用 json tag 重定义json中key的名字
	OrderID string `json:"orderID"`
	// 不使用json tag ，json中输出的key为 Grade
	Grade string
	// 忽略字段 "-" (无论有没有值, 都忽略)
	Username string `json:"-"`
	// 把int类型的数据展示为 string 类型，支持 字符串、浮点、整数及布尔类型转为string类型展示
	CourseID int `json:"courseID,string"`
	// 错误示例：把string类型作为int类型展示。不支持，无效的tag，仍然按照string类型输出
	UserID string `json:"userID,int"`
	// 忽略空值(值不为空, 不忽略)
	Desc string `json:"desc,omitempty"`
}

func BuyCourse(ctx *gin.Context, userID, courseID string) (c *BuyCourseInfo, err error) {
	// buy...

	// 购买成功后发送消息
	_ = SendRmqAfterBuyCourse(ctx, c)

	c = &BuyCourseInfo{
		Username: "xx",
		Grade:    "grade1",
		OrderID:  "7788",
		UserID:   userID,
		CourseID: 0,
		Desc:     fmt.Sprintf("this is random: %d", rand.Intn(1000)),
	}
	return c, nil
}

func GetUserInfo(ctx *gin.Context, userIDList []int) (u []user.User, err error) {
	u, err = user.GetUserByUserIDList(ctx, userIDList)
	if err != nil {
		return u, errors.WithMessagef(err, "models.GetUserByUserIDList() error, content: %v", userIDList)
	}

	return u, nil
}

func SendRmqAfterBuyCourse(ctx *gin.Context, content *BuyCourseInfo) (err error) {
	c, err := jsoniter.Marshal(content)
	if err != nil {
		return errors.WithMessagef(err, "content: %s", content.Grade)
	}
	msg, err := rmq.NewMessage("goweb", c)
	if err != nil {
		return errors.WithMessagef(err, "NewMessage() error, content: %s", c)
	}

	/*
		topic: 消息主题，通过 Topic 对不同的业务消息进行分类。
		tag: 消息标签，用于对某个 Topic 下的消息进行分类。生产者在发送消息时，已经指定消息的 Tag，消费者需根据已经指定的 Tag 来进行订阅。
		比如：订单消息和支付消息属于不同业务类型的消息，对应两个topic。其中订单消息根据商品品类以不同的 Tag 再进行细分，
		列如电器类、服装类、图书类等被各个不同的系统所接收。
		通过合理的使用 Topic 和 Tag，可以让业务结构清晰，更可以提高效率。
	*/

	msgID, err := msg.WithTag(content.Grade).Send(ctx)
	// msgID, err := msg.WithTag(content.Grade).WithDelay(rmq.Seconds30).Send()
	if err != nil {
		return errors.WithMessagef(err, "tag: %s, content: %s", content.Grade, c)
	}
	zlog.Debugf(ctx, "sent message id=%s, tag=%s", msgID, content.Grade)
	return nil
}
