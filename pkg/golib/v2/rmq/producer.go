package rmq

import (
	"context"
	"hash/fnv"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/zlog"
)

type Producer struct {
	conf     ProducerConf
	producer rocketmq.Producer
}

func newProducer(conf ProducerConf) (*Producer, error) {
	rander := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	nameServerAddr := getHostListByDns([]string{conf.NameServer})

	instance := instancePrefix + conf.Service
	options := []producer.Option{
		producer.WithInstanceName(instance + "-" + env.LocalIP + "-" + strconv.Itoa(os.Getpid()) + "-producer"),
		producer.WithNsResolver(primitive.NewPassthroughResolver(nameServerAddr)),
		producer.WithRetry(conf.Retry),
		producer.WithQueueSelector(&queueSelectorByShardingKey{Rander: rander}),
	}
	// 消息轨迹
	if conf.Trace {
		options = append(options, producer.WithTrace(&primitive.TraceConfig{
			TraceTopic: conf.TraceTopic,
			Access:     primitive.Local,
			Resolver:   primitive.NewPassthroughResolver(nameServerAddr),
		}))
	}

	if conf.Auth.AccessKey != "" && conf.Auth.SecretKey != "" {
		options = append(options, producer.WithCredentials(primitive.Credentials{
			AccessKey: conf.Auth.AccessKey,
			SecretKey: conf.Auth.SecretKey,
		}))
	}
	if conf.Timeout != 0 {
		options = append(options, producer.WithSendMsgTimeout(conf.Timeout))
	}
	prod, err := rocketmq.NewProducer(options...)
	if err != nil {
		logger.Error("failed to create producer",
			fields(zlog.String("ns", conf.NameServer), zlog.String("error", err.Error()))...)
		return nil, err
	}

	return &Producer{
		conf:     conf,
		producer: prod,
	}, nil
}

func (p *Producer) start() error {
	if p.producer == nil {
		return errors.Wrap(ErrRmqSvcInvalidOperation, "producer not initialized")
	}

	err := p.producer.Start()
	if err != nil {
		logger.Error("failed to start consumer", fields(zlog.String("error", err.Error()))...)
		return err
	}

	return nil
}

func (p *Producer) stop() error {
	return p.producer.Shutdown()
}

func (p *Producer) sendMessage(msgs ...*primitive.Message) (string, string, string, error) {
	res, err := p.producer.SendSync(context.Background(), msgs...)
	if err != nil {
		logger.Error("failed to send messages", fields(zlog.String("error", err.Error()))...)
		return "", "", "", err
	}
	return res.MessageQueue.String(), res.MsgID, res.OffsetMsgID, err
}

//ShardingKey hash方法
type queueSelectorByShardingKey struct {
	Rander *rand.Rand
}

func (q *queueSelectorByShardingKey) Select(msg *primitive.Message, queues []*primitive.MessageQueue) *primitive.MessageQueue {
	if msg.GetShardingKey() == "" {
		return queues[q.Rander.Intn(len(queues))]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(msg.GetShardingKey()))
	return queues[h.Sum32()%uint32(len(queues))]
}

type NmqResponse struct {
	TransID uint64 `mcpack:"_transid"`
	ErrNo   int    `mcpack:"_error_no" binding:"required"`
	ErrStr  string `mcpack:"_error_msg" binding:"required"`
}

// SendCmd 通过 Rmq 发送 Mcpack 格式的消息
// shard 消息分片参数, 按照传入值 hash 选择一个 MQ 分片并投递消息, 为空代表不做分片选择(会随机投递到一个MQ分片里)
func SendCmd(ctx *gin.Context, service string, cmd int64, topic string, product string, data map[string]interface{}, shard string) (resp NmqResponse, err error) {

	strCmd := strconv.FormatInt(cmd, 10)
	nmqMsg, transID := formatNmqMsg(ctx, strCmd, topic, product, data)

	body, err := mcpack.Marshal(nmqMsg) //数据mcpack编码
	if err != nil {
		return resp, err
	}

	msg, err := NewMessage(service, body)
	if err != nil {
		return resp, errors.WithMessagef(err, "NewMessage() error, body: %s", body)
	}
	msg = msg.WithTag(strCmd).WithKey(strconv.FormatUint(transID, 10))
	msg = msg.SetProperty("b", " ").SetProperty("s", " ") //mcpack格式数据要求
	// 如果消息需要按一定顺序投递, 则需要传入shard参数, 如uid或者orderid等值
	if shard != "" {
		msg = msg.WithShard(shard)
	}

	_, err = msg.Send(ctx)
	if err != nil {
		resp.ErrNo = -1
		resp.ErrStr = err.Error()
	} else {
		resp.ErrNo = 0
		resp.ErrStr = "OK"
		resp.TransID = transID
	}
	return resp, err
}
