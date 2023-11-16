package router

import (
	"permission/conf"
	"permission/controllers/mq"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/rmq"
)

// mq消费者回调入口
func MQ(g *gin.Engine) {
	// nmq 消费回调handler注册
	g.POST("/consume", mq.NmqTest)

	// rocketMQ 消费回调handler注册 , service 需要在 helpers/init.go 中注册（InitRocketMq中的 rmq.InitRmq）。
	// 一个应用尽可能用一个Topic，而消息子类型则可以用tags来标识。tags可以由应用自由设置，
	// 只有生产者在发送消息设置了tags，消费方在订阅消息时才可以利用tags通过broker做消息过滤
	// 建议不同格式的消息使用不同的Topic, Tag主要用于对相同格式消息的进一步拆分, 方便下游快速过滤出自己需要的消息。
	for _, consumeConf := range conf.RConf.Rmq.Consumer {
		// 初始化消费者
		if err := rmq.InitConsumer(consumeConf); err != nil {
			panic("register rmq[" + consumeConf.Service + "] error: " + err.Error())
		}
	}

	// rmq 消费回调 handler 注册
	// service 参数需要与 resource.yaml 中对应 consumer 配置的 service 字段对应
	err := rmq.StartConsumer(g, "consume-test", mq.BuyCourse)
	if err != nil {
		panic("Start consumer error: " + err.Error())
	}
	// start more consumers if needed.

	// kafka 消费回调handler注册
	// DemoSubClient := kafka.InitKafkaSub(g, conf.RConf.KafkaSub["demo"])
	// DemoSubClient.AddSubFunction(conf.RConf.KafkaSub["demo"].Topic, conf.RConf.KafkaSub["demo"].Group, mq.KafkaConsumerHandler, nil)
}
