package consumer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func NewKafkaConsumer(
	kafkaServerAddress string,
	kafkaConsumerID string,
	kafkaGroupID string,
) (*kafka.Consumer, error) {

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": kafkaServerAddress,
		"client.id":         kafkaConsumerID,
		"group.id":          kafkaGroupID,
	}

	c, err := kafka.NewConsumer(configMap)

	if err != nil {
		return nil, err
	}
	return c, nil
}
