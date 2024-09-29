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
	"log"
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

func (s *Suite) SubscribeToChat(ctx context.Context, userID string, messages chan<- *pb.Message) {
	stream, err := s.PubSubClient.Subscribe(context.Background(), &pb.SubscribeRequest{UserId: userID})
	if err != nil {
		return
	}

	defer func() {
		log.Println("closing connection")
		stream.CloseSend()
		close(messages)
	}()

	// Create a channel to receive stream messages or errors
	recvChan := make(chan *pb.Message)
	errChan := make(chan error)

	// Start a goroutine to receive messages from the stream
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					close(recvChan)
					return
				}
				errChan <- err
				return
			}
			recvChan <- msg
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Context canceled, stop receiving messages
			return
		case msg, ok := <-recvChan:
			if !ok {
				// Stream has been closed
				return
			}
			// Send the received message to the messages channel
			messages <- msg
		case err := <-errChan:
			// Handle any error from stream.Recv()
			_ = err // You can log or handle the error here
			return
		}
	}
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
