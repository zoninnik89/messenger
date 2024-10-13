package producer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	common "github.com/zoninnik89/commons"
	"log"
)

var (
	KafkaServerAddress = common.EnvString("KAFKA_SERVER_ADDRESS", "localhost:9092")
)

type MessageProducer struct {
	Producer *kafka.Producer
}

func NewKafkaProducer() (*MessageProducer, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": KafkaServerAddress,
		//"delivery.timeout.ms": "1",
		//"acks":                "all", //0-no ack, 1-leader, all,
		//"enable.idempotence":  "false",
	}
	p, err := kafka.NewProducer(configMap)

	if err != nil {
		return nil, err
	}

	return &MessageProducer{
		Producer: p,
	}, nil
}

func (p *MessageProducer) Publish(msg []byte, topic string, key []byte, deliveryChan chan kafka.Event) error {
	message := &kafka.Message{
		Value:          msg,
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
	}

	err := p.Producer.Produce(message, deliveryChan)

	if err != nil {
		return err
	}

	return nil
}

func (p *MessageProducer) DeliveryReport(deliveryChan chan kafka.Event) {
	for e := range deliveryChan {
		switch e.(type) {
		case *kafka.Message:
			e := <-deliveryChan
			msg := e.(*kafka.Message)

			if msg.TopicPartition.Error != nil {
				log.Println("message was not published")
			} else {
				log.Println("message published", msg.TopicPartition)
			}
		}
	}
}
