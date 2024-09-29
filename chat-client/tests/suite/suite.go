package tests

import (
	"context"
	"github.com/zoninnik89/messenger/chat-client/internal/config"
	pb "github.com/zoninnik89/messenger/common/api"
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
	Cfg              *config.Config
	ChatClientClient pb.ChatClientClient
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

	return ctx, &Suite{
		T:                t,
		Cfg:              cfg,
		ChatClientClient: pb.NewChatClientClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}

func (s *Suite) SubscribeToChat(ctx context.Context, userID string, messages chan<- *pb.Message) {
	stream, err := s.ChatClientClient.GetMessagesStream(context.Background(), &pb.GetMessagesStreamRequest{UserId: userID})
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

	_, err := s.ChatClientClient.SendMessage(ctx, &pb.SendMessageRequest{Message: msg})
	if err != nil {
		return err
	}
	
	return nil

}
