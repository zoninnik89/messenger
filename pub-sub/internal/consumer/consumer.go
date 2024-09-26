package consumer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	common "github.com/zoninnik89/messenger/common"
)

var (
	KafkaServerAddress = common.EnvString("KAFKA_SERVER_ADDRESS", "localhost:9092")
)

func NewKafkaConsumer(
	kafkaPort int,
	consumerID string,
	consumerGroupID string,
) (*kafka.Consumer, error) {

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": kafkaPort,
		"client.id":         consumerID,
		"group.id":          consumerGroupID,
	}

	c, err := kafka.NewConsumer(configMap)

	if err != nil {
		return nil, err
	}
	return c, nil
}
