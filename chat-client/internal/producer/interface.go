package producer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ClickProducerInterface interface {
	Publish(msg string, topic string, key []byte, deliveryChan chan kafka.Event) error
	DeliveryReport(deliveryChan chan kafka.Event)
}
