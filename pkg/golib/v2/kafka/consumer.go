package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"permission/pkg/golib/v2/base"
	secret "permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/middleware"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const kafkaSubPrefix = "@@kafkasub."
const BodyKey string = "KafkaMsg"
const msgKey string = "__KafkaMsg"
const defaultKafkaVersion = "2.2.1"

type ConsumeConfig struct {
	Service string   `yaml:"service"`
	Version string   `yaml:"version"`
	Brokers []string `yaml:"brokers"`
	Balance string   `yaml:"balance"`
	SASL    sasl     `yaml:"sasl"`
	Newest  bool
	Topic   []string `yaml:"topic"`
	Group   string   `yaml:"group"`
}
type sasl struct {
	Enable    bool                 `yaml:"enable"`
	Handshake bool                 `yaml:"handshake"`
	User      string               `yaml:"user"`
	Password  string               `yaml:"password"`
	Mechanism sarama.SASLMechanism `yaml:"mechanism"`
}
type SubClient struct {
	Brokers  []string
	Version  sarama.KafkaVersion
	Strategy sarama.BalanceStrategy
	sasl     sasl
	g        *gin.Engine
	cancel   context.CancelFunc
}

func (conf *ConsumeConfig) CheckConfig() {
	secret.CommonSecretChange(kafkaSubPrefix, *conf, conf)

	if conf.Version == "" {
		conf.Version = defaultKafkaVersion
	}
}

func InitKafkaSub(g *gin.Engine, subConf ConsumeConfig) *SubClient {
	subConf.CheckConfig()
	v, err := sarama.ParseKafkaVersion(subConf.Version)
	if err != nil {
		panic("Error parsing Kafka version: " + err.Error())
	}

	var s sarama.BalanceStrategy
	// 默认分区策略是 range
	if subConf.Balance == "roundrobin" {
		s = sarama.BalanceStrategyRoundRobin
	} else {
		s = sarama.BalanceStrategyRange
	}

	return &SubClient{
		Version:  v,
		g:        g,
		Brokers:  subConf.Brokers,
		Strategy: s,
		sasl:     subConf.SASL,
	}
}

type kafkaHandler func(*gin.Context) error

type ConsumerOption struct {
	ConsumerFromNewest bool
	IsRawData          bool
}

func (c *SubClient) AddSubFunction(topics []string, groupID string, handler kafkaHandler, opts *ConsumerOption) {
	config := sarama.NewConfig()
	config.Version = c.Version
	config.Consumer.Group.Rebalance.Strategy = c.Strategy

	if opts == nil || opts.ConsumerFromNewest == true {
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	} else {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	if c.sasl.Enable {
		config.Net.SASL.Mechanism = c.sasl.Mechanism
		config.Net.SASL.User = c.sasl.User
		config.Net.SASL.Password = c.sasl.Password
		config.Net.SASL.Handshake = c.sasl.Handshake
		config.Net.SASL.Enable = true
	}
	// config.Consumer.Return.Errors = true
	consumerGroup, err := sarama.NewConsumerGroup(c.Brokers, groupID, config)
	if err != nil {
		panic("NewConsumerGroup error: " + err.Error())
	}

	consumerHandler := &ConsumerGroup{
		handler: handler,
		Client:  c,
		Ready:   make(chan bool),
	}
	if opts != nil {
		consumerHandler.IsRawData = opts.IsRawData
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go func() {
		for {

			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := consumerGroup.Consume(ctx, topics, consumerHandler); err != nil {
				zlog.Warn(nil, "Error from consumer: ", err.Error())
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumerHandler.Ready = make(chan bool, 0)

		}
	}()

	// Await till the consumer has been set up
	<-consumerHandler.Ready
	return
}

func (c *SubClient) CloseConsumer() {
	c.cancel()
}

type ConsumerGroup struct {
	Ready     chan bool
	handler   func(*gin.Context) error
	Client    *SubClient
	IsRawData bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *ConsumerGroup) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *ConsumerGroup) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *ConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		err := c.HandleMessage(message)
		if err != nil {
			continue
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func (c *ConsumerGroup) HandleMessage(message *sarama.ConsumerMessage) (err error) {
	ctx := gin.CreateNewContext(c.Client.g)
	customCtx := gin.CustomContext{
		Handle:    c.handler,
		Desc:      message.Topic,
		Type:      "Kafka",
		StartTime: time.Now(),
	}
	ctx.CustomContext = customCtx

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to consume message, %v", r)
			gin.CustomerErrorLog(ctx, err.Error(), true, map[string]string{
				"handle": ctx.CustomContext.HandlerName(),
				"desc":   ctx.CustomContext.Desc,
			})
		}

		gin.RecycleContext(c.Client.g, ctx)
	}()

	var body Body
	if c.IsRawData {
		body.Msg = message.Value
	} else if err := json.Unmarshal(message.Value, &body); err == nil {
		// kafka 消息发送的时候默认包装了一层msg，这里做个兼容。
		if body.Msg == nil {
			if err := json.Unmarshal(message.Value, &body.Msg); err != nil {
				zlog.Warn(ctx, "got unexpected value")
			}
		}
	} else {
		body.Msg = message.Value
	}

	middleware.UseMetadata(ctx)

	ctx.Set(BodyKey, body.Msg)
	ctx.Set(msgKey, message)
	err = c.handler(ctx)

	// info 日志里打印出partition
	zlog.AddNotice(ctx, "partition", message.Partition)

	ctx.CustomContext.Error = err
	ctx.CustomContext.EndTime = time.Now()
	loggerAfterHandle(ctx)
	return err
}

func loggerAfterHandle(ctx *gin.Context) {
	customCtx := ctx.CustomContext
	cost := utils.GetRequestCost(customCtx.StartTime, customCtx.EndTime)
	if customCtx.Error != nil {
		base.StackLogger(ctx, customCtx.Error)
	}

	var fields []zlog.Field

	errMsg := "<nil>"
	if customCtx.Error != nil {
		errMsg = customCtx.Error.Error()
	}
	fields = append(fields,
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("handle", customCtx.HandlerName()),
		zlog.String("type", customCtx.Type),
		zlog.Float64("cost", cost),
		zlog.String("desc", customCtx.Desc),
		zlog.String("error", errMsg),
	)

	// 用户自定义notice
	notices := zlog.GetCustomerKeyValue(ctx)

	for k, v := range notices {
		fields = append(fields, zlog.Reflect(k, v))
	}

	zlog.InfoLogger(ctx, "end", fields...)
}

func GetKafkaMsg(ctx *gin.Context) (msg interface{}, exist bool) {
	msg, exist = ctx.Get(BodyKey)
	return msg, exist
}

func GetMessage(ctx *gin.Context) (message *sarama.ConsumerMessage, ok bool) {
	msg, ok := ctx.Get(msgKey)
	if !ok {
		zlog.Error(ctx, "get empty kafka consumer message")
		return nil, false
	}

	message, ok = msg.(*sarama.ConsumerMessage)
	if !ok {
		zlog.Error(ctx, "kafka consumer message: invalid format")
		return nil, false
	}

	return message, true
}
