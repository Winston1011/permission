package data_test

import (
	"encoding/json"
	"testing"
	"time"

	"permission/helpers"

	"permission/pkg/golib/v2/nmq"
	"permission/pkg/golib/v2/rmq"
)

func TestRMQ_Send(t *testing.T) {
	content := make(map[string]string)
	content["orderID"] = "1234567890"
	content["data"] = "test"
	body, _ := json.Marshal(content)

	// service名称需要与resource.yaml中的服务名称对应
	msg, err := rmq.NewMessage("rocketmq-test", body)
	if err != nil {
		t.Error()
		return
	}

	/*
		topic: 消息主题，通过 Topic 对不同的业务消息进行分类。
		tag: 消息标签，用于对某个 Topic 下的消息进行分类。生产者在发送消息时，已经指定消息的 Tag，消费者需根据已经指定的 Tag 来进行订阅。
		比如：订单消息和支付消息属于不同业务类型的消息，对应两个topic。其中订单消息根据商品品类以不同的 Tag 再进行细分，
		列如电器类、服装类、图书类等被各个不同的系统所接收。
		通过合理的使用 Topic 和 Tag，可以让业务结构清晰，更可以提高效率。
	*/

	msg = msg.WithTag("tagA")
	// 跟消息正文无关的辅助字段, 可以放在Header中传递
	msg = msg.WithHeader("Time", time.Now().String())
	msgID, err := msg.Send(ctx)
	if err != nil {
		t.Error("[TestRMQ_Send] send error: ", err.Error())
		return
	}

	t.Logf("sent message id=%s", msgID)
}

func TestNmqSendExample(t *testing.T) {
	body := map[string]interface{}{
		"courseID":  "12xxx",
		"studentID": "34yyy",
	}

	topic, product := "goweb", "inf"
	resp, err := nmq.SendCmd(ctx, 10001, topic, product, body)
	if err != nil {
		t.Error("nmq send error:", err.Error())
		return
	}

	t.Logf("nmq send succ, resp: %+v", resp)
}

func TestKafka_PubMessage(t *testing.T) {
	// 直接发送字符串
	if err := helpers.DemoPubClient.Pub(ctx, "topic1", "hello1"); err != nil {
		t.Error("[TestKafka_PubMessage] error: ", err.Error())
		return
	}

	// 直接发送结构体，对方收到的是该结构体的json encode后的[]byte
	msg2 := struct {
		UserName string
		Age      int
	}{
		UserName: "jay",
		Age:      22,
	}
	if err := helpers.DemoPubClient.Pub(ctx, "topic2", msg2); err != nil {
		t.Error("[TestKafka_PubMessage] error: ", err.Error())
		return
	}
}
