package producer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type Producer struct {
	Producer *kafka.Producer
}

func NewKafkaProducer() *Producer {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}
	p, err := kafka.NewProducer(configMap)

	if err != nil {
		log.Println(err.Error())
	}

	return &Producer{
		Producer: p,
	}
}

func (p *Producer) Publish(msg []byte, topic string, key []byte, deliveryChan chan kafka.Event) error {
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

func (p *Producer) DeliveryReport(deliveryChan chan kafka.Event) {
	for e := range deliveryChan {
		switch e.(type) {
		case *kafka.Message:
			e := <-deliveryChan
			msg := e.(*kafka.Message)

			if msg.TopicPartition.Error != nil {
				log.Println("Message was not published")
			} else {
				log.Println("Message published", msg.TopicPartition)
			}
		}
	}
}
