package suite

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/config"
	producer "github.com/zoninnik89/messenger/pub-sub/tests/suite/mock-kafka-producer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"strconv"
	"testing"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient pb.AuthServiceClient
	Queue      *producer.Producer
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local.yaml")
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.NewClient(grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	p := producer.NewKafkaProducer()

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: pb.NewAuthServiceClient(cc),
		Queue:      p,
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}

func (s *Suite) ProduceMessage(msg *pb.Message) {
	log.Printf("Produce message request received")

	serializedMessage, err := proto.Marshal(msg)
	if err != nil {
		log.Fatalf("protobuf serialization error: %s", err)
	}

	deliveryChan := make(chan kafka.Event)
	err = s.Queue.Publish(serializedMessage, "messages", nil, deliveryChan)

	if err != nil {
		log.Println("error publishing click event in Kafka", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Println("message was not published", "error", m.TopicPartition.Error)
	} else {
		log.Println("message successfully published", "message", m.TopicPartition, "time", m.Timestamp.String())
	}
	close(deliveryChan)

}
