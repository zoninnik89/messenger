package consumer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	common "github.com/zoninnik89/commons"
	"log"
)

var (
	KafkaServerAddress = common.EnvString("KAFKA_SERVER_ADDRESS", "localhost:9092")
)

func NewKafkaConsumer() *kafka.Consumer {

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": KafkaServerAddress,
		"client.id":         "aggregator-consumer",
		"group.id":          "aggregator-group",
	}

	c, err := kafka.NewConsumer(configMap)

	if err != nil {
		log.Println("error consumer", err.Error())
	}
	return c
}
