package rmq

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/middleware"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

type pushConsumer struct {
	conf     ConsumerConf
	consumer rocketmq.PushConsumer
	engine   *gin.Engine
}

func newPushConsumer(conf ConsumerConf) (*pushConsumer, error) {
	instance := instancePrefix + conf.Service
	var c consumer.MessageModel
	if conf.Broadcast {
		instance = instance + "-consumer"
		c = consumer.BroadCasting
	} else {
		instance = instance + "-" + env.LocalIP + "-" + strconv.Itoa(os.Getpid()) + "-consumer"
		c = consumer.Clustering
	}

	nameServerAddr := getHostListByDns([]string{conf.NameServer})
	options := []consumer.Option{
		consumer.WithInstance(instance),
		consumer.WithGroupName(conf.Group),
		consumer.WithAutoCommit(true),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(nameServerAddr)),
		consumer.WithConsumerOrder(conf.Orderly),
		consumer.WithConsumeMessageBatchMaxSize(conf.Batch),
		consumer.WithMaxReconsumeTimes(int32(conf.Retry)),
		consumer.WithStrategy(consumer.AllocateByAveragely),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithConsumerModel(c),
		consumer.WithSuspendCurrentQueueTime(conf.RetryInterval), // 顺序消费设置 失败重试间隔
	}
	// 配置开启了消息轨迹
	if conf.Trace {
		options = append(options, consumer.WithTrace(&primitive.TraceConfig{
			TraceTopic: conf.TraceTopic,
			Access:     primitive.Local,
			Resolver:   primitive.NewPassthroughResolver(nameServerAddr),
		}))
	}

	if conf.Auth.AccessKey != "" && conf.Auth.SecretKey != "" {
		options = append(options, consumer.WithCredentials(primitive.Credentials{
			AccessKey: conf.Auth.AccessKey,
			SecretKey: conf.Auth.SecretKey,
		}))
	}

	con, err := rocketmq.NewPushConsumer(options...)
	if err != nil {
		logger.Error("failed to create consumer", fields(zlog.String("error", err.Error()))...)
		return nil, err
	}

	return &pushConsumer{
		conf:     conf,
		consumer: con,
	}, nil
}

func (c *pushConsumer) start(callback MessageCallback) (err error) {
	err = c.subscribe(callback)
	if err != nil {
		return err
	}

	return c.consumer.Start()
}

func (c *pushConsumer) stop() error {
	return c.consumer.Shutdown()
}

func (c *pushConsumer) subscribe(callback MessageCallback) (err error) {
	if callback == nil {
		return errors.Wrap(ErrRmqSvcInvalidOperation, "nil callback")
	}

	cb := func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, m := range msgs {
			if ctx.Err() != nil {
				logger.Error("stop consume cause ctx cancelled", fields(
					zlog.String("service", c.conf.Service),
					zlog.String("error", err.Error()))...)
				return consumer.SuspendCurrentQueueAMoment, ctx.Err()
			}
			if err := c.call(callback, m); err != nil {
				return consumer.SuspendCurrentQueueAMoment, nil
			}
		}
		return consumer.ConsumeSuccess, nil
	}

	var expr string
	if len(c.conf.Tags) == 0 {
		expr = "*"
	} else if len(c.conf.Tags) == 1 {
		expr = c.conf.Tags[0]
	} else {
		expr = strings.Join(c.conf.Tags, "||")
	}

	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: expr,
	}

	err = c.consumer.Subscribe(c.conf.Topic, selector, cb)
	if err != nil {
		logger.Error("failed to subscribe", fields(zlog.String("error", err.Error()))...)
		return err
	}

	return nil
}

func (c *pushConsumer) call(fn MessageCallback, m *primitive.MessageExt) (err error) {
	ctx := gin.CreateNewContext(c.engine)
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to consume message, %v", r)
			gin.CustomerErrorLog(ctx, err.Error(), true, nil)
		}
		gin.RecycleContext(c.engine, ctx)
	}()

	produceTime := time.Unix(m.BornTimestamp/1000, m.BornTimestamp%1000*int64(time.Millisecond))
	mw := &messageWrapper{
		msg:   &m.Message,
		msgID: m.MsgId,
		time:  produceTime,
		retry: int(m.ReconsumeTimes),
	}

	if m.Message.GetTags() != "" && m.Message.GetKeys() != "" {
		pack, err := mcpack.Decode(m.Message.Body)
		if err == nil {
			logid := fmt.Sprintf("%v", pack.Get(packKeyLogID))
			callerUri := fmt.Sprintf("%v", pack.Get(packKeyCallerURI))
			ctx.Set(zlog.ContextKeyLogID, logid)
			if callerUri != "" && callerUri != "\x00" {
				ctx.Set(packKeyCallerURI, callerUri)
			}
		}
	}

	// mq 中取得的 "X-Zyb-Ctx-" 开头 headers，设置到 ctx
	utils.SetTransportHeader(ctx, mw.GetCtxHeaders())
	middleware.UseMetadata(ctx)

	err = fn(ctx, mw)
	consumeResult := "consume message success"
	ralCode := 0
	if err != nil {
		ralCode = -1
		consumeResult = err.Error()
	}

	// 用户自定义notice
	end := time.Now()
	var fields []zlog.Field
	for k, v := range zlog.GetCustomerKeyValue(ctx) {
		fields = append(fields, zlog.Reflect(k, v))
	}

	// not modifiy original msg body
	var bodyLog bytes.Buffer
	if len(m.Body) > 100 {
		bodyLog.Write(m.Body[0:100])
		bodyLog.WriteString("...")
	} else {
		bodyLog.Write(m.Body)
	}
	fields = append(fields,
		zlog.String("service", c.conf.Service),
		zlog.String("addr", c.conf.NameServer),
		zlog.String("method", "consume"),
		zlog.String("topic", m.Topic),
		zlog.String("message", bodyLog.String()),
		zlog.String("queue", m.Queue.String()),
		zlog.String("msgID", m.MsgId),
		// zlog.String("offsetMsgID", m.OffsetMsgId),
		zlog.String("msgkey", m.GetKeys()),
		zlog.String("tags", m.GetTags()),
		zlog.String("shard", m.GetShardingKey()),
		zlog.String("headers", fmtHeaders(&m.Message, HeaderPre)),
		// zlog.String("ctxHeaders", fmtHeaders(&m.Message, utils.ZYBTransportHeader)),
		zlog.Int("size", len(m.Body)),
		zlog.Int("ralCode", ralCode),
		zlog.String("response", consumeResult),
		zlog.String("retry", strconv.Itoa(int(m.ReconsumeTimes))+"/"+strconv.Itoa(c.conf.Retry)),
		zlog.Int64("delay", start.UnixNano()/int64(time.Millisecond)-m.BornTimestamp),
		zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		zlog.Float64("cost", utils.GetRequestCost(start, end)),
	)

	logger.Info("rmq-consume", contextFields(ctx, fields...)...)
	if err != nil {
		logger.Error("failed to consume message: "+err.Error(), contextFields(ctx, fields...)...)
	}

	return err
}
