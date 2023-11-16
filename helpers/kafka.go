package helpers

import (
	"permission/conf"

	"permission/pkg/golib/v2/kafka"
)

var DemoPubClient *kafka.PubClient

func InitKafkaProducer() {
	DemoPubClient = kafka.InitKafkaPub(conf.RConf.KafkaPub["demo"])

	if DemoPubClient == nil {
		panic("init redis failed!")
	}
}

func CloseKafkaProducer() {
	_ = DemoPubClient.CloseProducer()
}
