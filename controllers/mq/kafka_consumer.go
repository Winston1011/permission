package mq

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/kafka"
	"permission/pkg/golib/v2/zlog"
)

// 处理方法
func KafkaConsumerHandler(ctx *gin.Context) error {
	message, ok := kafka.GetMessage(ctx)
	if !ok {
		// 忽略本条消息
		return nil
	}

	// 同一个 consumerGroup 可以消费多个topic
	switch message.Topic {
	case "topic1":
		// do something

		/*
			注意 message.Value 是原始发出去的消息,如果是使用 golib 发送的消息，底层默认加了一层 Msg，比如：
			如果发送的字符串是 "hello" ， 那么 message.Value 是 {"msg": "hello"}
			这个是由于生产者历史代码设计有问题，如果不想使用这个功能，可以在生产的配置中增加 rawMessage: true
		*/
		zlog.Debug(ctx, "[KafkaConsumerHandler] demo1 GetKafkaMsg: ", string(message.Value))
	case "topic2":
		var user struct {
			UserName string
			Age      int
		}

		if err := json.Unmarshal(message.Value, &user); err != nil {
			zlog.Error(ctx, "Unmarshal error: ", err.Error())
		}
		zlog.Debugf(ctx, "[KafkaConsumerHandler] topic[demo2] username: %s, age: %d ", user.UserName, user.Age)
	}

	return nil
}
