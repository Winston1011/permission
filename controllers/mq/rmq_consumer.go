package mq

import (
	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/rmq"
	"permission/pkg/golib/v2/zlog"
)

/*
	关于消费幂等
	RocketMQ无法避免消息重复（Exactly-Once），所以如果业务对消费重复非常敏感，务必要在业务层面进行去重处理。
	可以借助关系数据库进行去重。首先需要确定消息的唯一键，可以是msgId，也可以是消息内容中的唯一标识字段，例如订单Id等。
	在消费之前判断唯一键是否在关系数据库中存在。如果不存在则插入，并消费，否则跳过。（实际过程要考虑原子性问题，判断是否存在可以尝试插入，如果报主键冲突，则插入失败，直接跳过）
	msgId一定是全局唯一标识符，但是实际使用中，可能会存在相同的消息有两个不同msgId的情况（消费者主动重发、因客户端重投机制导致的重复等），这种情况就需要使业务字段进行重复消费。
*/
func BuyCourse(ctx *gin.Context, msg rmq.Message) error {
	// do something.
	zlog.Info(ctx, "got message id=", msg.GetID(), " tag=", msg.GetTag(), " header=", msg.GetHeader("Time"))
	content := msg.GetContent()

	// do something

	zlog.Debug(ctx, "content is ", string(content))

	// 消费成功需要return nil, 消费失败需要return err. 失败消息后续会重试消费
	return nil
}
