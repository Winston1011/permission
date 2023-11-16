package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"permission/pkg/golib/v2/env"
	secret "permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

type ProducerConfig struct {
	Service  string `yaml:"service"`
	Addr     string `yaml:"addr"`
	Version  string `yaml:"version"`
	IsRawMsg bool   `yaml:"rawMsg"`
	SASL     sasl   `yaml:"sasl"`
	TLS      struct {
		Enable                bool   `yaml:"enable"`
		CA                    string `yaml:"ca"`
		Cert                  string `yaml:"cert"`
		Key                   string `yaml:"key"`
		InsecureSkipTLSVerify bool   `yaml:"insecure_skip_tls_verify"`
	} `yaml:"tls"`
}
type PubClient struct {
	Conf     ProducerConfig
	producer sarama.SyncProducer
}

type Body struct {
	Msg interface{}
}

const kafkaPubPrefix = "@@kafkapub."

func (conf *ProducerConfig) GetKafkaConfig() (*sarama.Config, error) {
	secret.CommonSecretChange(kafkaPubPrefix, *conf, conf)

	if conf.Version == "" {
		conf.Version = defaultKafkaVersion
	}

	defaultConfig := sarama.NewConfig()
	v, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		return nil, err
	}
	defaultConfig.Version = v
	if conf.SASL.Enable {
		defaultConfig.Net.SASL.Enable = true
		defaultConfig.Net.SASL.Handshake = conf.SASL.Handshake
		defaultConfig.Net.SASL.Mechanism = conf.SASL.Mechanism
		defaultConfig.Net.SASL.User = conf.SASL.User
		defaultConfig.Net.SASL.Password = conf.SASL.Password
	}
	if conf.TLS.Enable {
		defaultConfig.Net.TLS.Enable = true
		defaultConfig.Net.TLS.Config = &tls.Config{
			RootCAs:            x509.NewCertPool(),
			InsecureSkipVerify: conf.TLS.InsecureSkipTLSVerify,
		}
		if conf.TLS.CA != "" {
			ca, err := os.ReadFile(conf.TLS.CA)
			if err != nil {
				panic("kafka pub CA error: %v" + err.Error())
			}
			defaultConfig.Net.TLS.Config.RootCAs.AppendCertsFromPEM(ca)
		}
	}
	defaultConfig.Producer.Return.Successes = true

	return defaultConfig, nil
}

func InitKafkaPub(conf ProducerConfig) *PubClient {
	saramaConfig, err := conf.GetKafkaConfig()
	if err != nil {
		panic("kafka pub version error: %v" + err.Error())
	}

	addrs := strings.Split(strings.TrimSpace(conf.Addr), ",")
	producer, err := sarama.NewSyncProducer(addrs, saramaConfig)
	if err != nil {
		panic("kafka pub new producer error: %v" + err.Error())
	}

	c := &PubClient{
		Conf:     conf,
		producer: producer,
	}
	return c
}

func (client *PubClient) CloseProducer() error {
	if client.producer != nil {
		return client.producer.Close()
	}
	return nil
}

func (client *PubClient) Pub(ctx *gin.Context, topic string, msg interface{}) error {
	if client.producer == nil {
		return errors.New("kafka producer not init")
	}

	// todo 低版本中发送消息的时候默认包了一层。这是个大坑，消费者不一定是go服务，可能不好解析msg字段。
	//  建议配置中增加 rawMessage: true 来取消这层封装
	var toSendBody interface{}
	toSendBody = Body{
		Msg: msg,
	}
	if client.Conf.IsRawMsg {
		toSendBody = msg
	}

	body, err := json.Marshal(toSendBody)
	if err != nil {
		return err
	}

	start := time.Now()
	kafkaMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(body)}
	partition, offset, err := client.producer.SendMessage(kafkaMsg)
	end := time.Now()

	ralCode := 0
	infoMsg := "kafka pub success"
	if err != nil {
		ralCode = -1
		infoMsg = err.Error()
		zlog.ErrorLogger(ctx, "kafka pub error: "+infoMsg, zlog.String(zlog.TopicType, zlog.LogNameModule))
	}

	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("requestId", zlog.GetRequestID(ctx)),
		zlog.String("localIp", env.LocalIP),
		zlog.String("remoteAddr", client.Conf.Addr),
		zlog.String("service", client.Conf.Service),
		zlog.Int32("partition", partition),
		zlog.Int64("offset", offset),
		zlog.Int("ralCode", ralCode),
		zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		zlog.Float64("cost", utils.GetRequestCost(start, end)),
	}

	zlog.InfoLogger(ctx, infoMsg, fields...)

	return nil
}
