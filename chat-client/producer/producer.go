package producer

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	common "github.com/zoninnik89/commons"
	"log"
)

var (
	KafkaServerAddress = common.EnvString("KAFKA_SERVER_ADDRESS", "localhost:9092")
)

type ClickProducer struct {
	Producer *kafka.Producer
}

func NewKafkaProducer() *ClickProducer {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": KafkaServerAddress,
		//"delivery.timeout.ms": "1",
		//"acks":                "all", //0-no ack, 1-leader, all,
		//"enable.idempotence":  "false",
	}
	p, err := kafka.NewProducer(configMap)

	if err != nil {
		log.Println(err.Error())
	}

	return &ClickProducer{
		Producer: p,
	}
}

func (p *ClickProducer) Publish(msg string, topic string, key []byte, deliveryChan chan kafka.Event) error {
	message := &kafka.Message{
		Value:          []byte(msg),
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
	}

	err := p.Producer.Produce(message, deliveryChan)

	if err != nil {
		return err
	}

	return nil
}

func (p *ClickProducer) DeliveryReport(deliveryChan chan kafka.Event) {
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

//func (p *ClickProducer) Flush(timeoutMs int) int {
//	termChan := make(chan bool)
//
//	d, _ := time.ParseDuration(fmt.Sprintf("%dms", timeoutMs))
//	tEnd := time.Now().Add(d)
//	for p.producer.Len() > 0 {
//		remain := tEnd.Sub(time.Now()).Seconds()
//		if remain <= 0.0 {
//			return p.producer.Len()
//		}
//
//		p.producer.handle.eventPoll(p.producer.events,
//			int(math.Min(100, remain*1000)), 1000, termChan)
//	}
//
//	return 0
//}
