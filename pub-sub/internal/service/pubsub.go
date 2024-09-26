package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/pub-sub/internal/logging"
	"github.com/zoninnik89/messenger/pub-sub/internal/storage"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"strings"

	pb "github.com/zoninnik89/messenger/common/api"
	"go.uber.org/zap"
)

type PubSubService struct {
	Chats      *storage.AsyncMap
	Logger     *zap.SugaredLogger
	chanBuffer int
}

func NewPubSubService(chanBuffer int) *PubSubService {
	return &PubSubService{Chats: storage.NewAsyncMap(), Logger: logging.GetLogger().Sugar(), chanBuffer: chanBuffer}
}

var (
	ErrChatNotExists         = errors.New("chat not exists")
	ErrNoChatSubscribers     = errors.New("no chat subscribers")
	ErrClientNotFound        = errors.New("client not found")
	SuccessfullyConsumedResp = "message was sent to all recipients"
)

func (p *PubSubService) Subscribe(
	chatID string,
	stream pb.PubSubService_SubscribeServer,
) error {
	var op = "service.Subscribe"

	p.Logger.Infow("op", op, "chat ID", chatID)
	channel := make(chan *pb.MessageResponse, p.chanBuffer)
	client := &types.Client{
		MessageChannel: &channel,
	}
	p.Chats.Add(chatID, client)

	for {
		select {
		case msg := <-*client.MessageChannel:
			p.Logger.Infow("received message from client channel", "message", msg.Message)
			if err := stream.Send(msg); err != nil {
				p.Logger.Errorw("error sending message to client", "err", err)
				p.removeClient(chatID, client) // Removing the client from the Async map
				return fmt.Errorf("%s: %w", op, err)
			}
		case <-stream.Context().Done():
			p.Logger.Infow("Client disconnected from chat", "chat", chatID)
			p.removeClient(chatID, client) // Removing the client from the Async map
			return nil
		}
	}
}

func (p *PubSubService) ConsumeMessage(ctx context.Context, consumer *kafka.Consumer) (string, error) {
	var op = "service.ConsumeMessage"

	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		p.Logger.Fatalw("failed to read message", "err", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	msgSlice := strings.Split(string(msg.Value), ",")
	chatID, senderID, messageID, messageText, sentTime := msgSlice[0], msgSlice[1], msgSlice[2], msgSlice[3], msgSlice[4]

	clients, err := p.Chats.Get(chatID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, ErrChatNotExists)
	}

	// If there are no available recipients
	if clients.Size() == 0 {
		return "", fmt.Errorf("%s: %w", op, ErrNoChatSubscribers)
	}

	// Send the message to all clients
	for client := range clients.Store {
		p.Logger.Infow("sending message to client", "chatID", chatID, "senderID", senderID, "messageID", messageID, "messageText", messageText, "sentTime", sentTime)
		*client.MessageChannel <- &pb.MessageResponse{Message: &pb.Message{
			SenderId:    senderID,
			MessageId:   messageID,
			MessageText: messageText,
			SentTs:      sentTime,
		}}
	}
	p.Logger.Infow("message successfully consumed", "chatID", chatID, "messageID", messageID)

	return SuccessfullyConsumedResp, nil
}

func (p *PubSubService) removeClient(chat string, client *types.Client) {
	var op = "service.RemoveClient"

	err := p.Chats.Remove(chat, client)
	if err != nil {
		p.Logger.Errorw("op", op, "err", ErrClientNotFound)

	}
}
