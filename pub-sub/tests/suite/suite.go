package suite

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/config"
	producer "github.com/zoninnik89/messenger/pub-sub/tests/suite/mock-kafka-producer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg          *config.Config
	PubSubClient pb.PubSubServiceClient
	Queue        *producer.Producer
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	//t.Parallel()

	cfg := config.MustLoadByPath("../config/local.yaml")
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.NewClient(grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil
	}

	p := producer.NewKafkaProducer()

	return ctx, &Suite{
		T:            t,
		Cfg:          cfg,
		PubSubClient: pb.NewPubSubServiceClient(cc),
		Queue:        p,
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}

func (s *Suite) SubscribeToChat(ctx context.Context, chat string, messages chan<- *pb.Message) {
	stream, err := s.PubSubClient.Subscribe(context.Background(), &pb.SubscribeRequest{ChatId: chat})
	if err != nil {
		return
	}

	// Continuously receive messages from the stream
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			// Server closed connection
			break
		}
		if err != nil {
			return
		}

		// Log the received message
		messages <- msg.Message
	}

	close(messages)
}

func (s *Suite) SendMessage(
	ctx context.Context,
	messageID string,
	chatID string,
	senderID string,
	messageText string,
) error {
	// publish message to the chat

	msg := &pb.Message{
		MessageId:   messageID,
		SenderId:    senderID,
		ChatId:      chatID,
		MessageText: messageText,
		SentTs:      strconv.FormatInt(time.Now().Unix(), 10),
	}

	deliveryChan := make(chan kafka.Event)
	err := s.Queue.Publish(msg, "messages", nil, deliveryChan)

	if err != nil {
		return err
	}

	e := <-deliveryChan
	dmsg := e.(*kafka.Message)

	if dmsg.TopicPartition.Error != nil {
		return dmsg.TopicPartition.Error
	}

	close(deliveryChan)
	return nil
}
