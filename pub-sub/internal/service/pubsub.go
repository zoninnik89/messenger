package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/logging"
	"github.com/zoninnik89/messenger/pub-sub/internal/storage"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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
	ErrChatNotExists       = errors.New("chat not exists")
	ErrNoChatSubscribers   = errors.New("no chat subscribers")
	ErrClientNotFound      = errors.New("client not found")
	ErrNoMessageID         = errors.New("no message ID")
	ErrMessageMissingField = errors.New("message misses one of the fields")
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
			p.Logger.Infow("op", op, "received message from client channel", "message", msg.Message)
			if err := stream.Send(msg); err != nil {
				p.Logger.Errorw("op", op, "error sending message to client", "err", err)
				p.removeClient(chatID, client) // Removing the client from the Async map
				return fmt.Errorf("%s: %w", op, err)
			}
		case <-stream.Context().Done():
			p.Logger.Infow("op", op, "Client disconnected from chat", "chat", chatID)
			p.removeClient(chatID, client) // Removing the client from the Async map
			return nil
		}
	}
}

func (p *PubSubService) ConsumeMessage(ctx context.Context, consumer *kafka.Consumer) (string, error) {
	var op = "service.ConsumeMessage"

	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		p.Logger.Fatalw("op", op, "failed to read message", "err", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var deserializedMessage pb.Message
	if err := proto.Unmarshal(msg.Value, &deserializedMessage); err != nil {
		p.Logger.Fatalw("op", op, "failed to unmarshal message", "err", err)
	}

	messageID := deserializedMessage.GetMessageId()

	p.Logger.Infow("op", op, "validating message:", "message", messageID)

	if err := p.validateMessage(&deserializedMessage); err != nil {
		p.Logger.Errorw("op", op, "invalid message", "err", err)
		return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, err)
	}

	chatID := deserializedMessage.GetChatId()
	senderID := deserializedMessage.GetSenderId()
	messageText := deserializedMessage.GetMessageText()
	sentTime := deserializedMessage.GetSentTs()

	clients, err := p.Chats.Get(chatID)
	if err != nil {
		return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, ErrChatNotExists)
	}

	// If there are no available recipients
	if clients.Size() <= 1 {
		return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, ErrNoChatSubscribers)
	}

	// Send the message to all clients
	for client := range clients.Store {
		p.Logger.Infow("sending message to client", "chatID", chatID, "senderID", senderID, "messageID", messageID, "messageText", messageText, "sentTime", sentTime)
		*client.MessageChannel <- &pb.MessageResponse{Message: &pb.Message{
			ChatId:      chatID,
			SenderId:    senderID,
			MessageId:   messageID,
			MessageText: messageText,
			SentTs:      sentTime,
		}}
	}
	p.Logger.Infow("op", op, "message successfully consumed", "chatID", chatID, "messageID", messageID)

	return messageID, nil
}

func (p *PubSubService) removeClient(chatID string, client *types.Client) {
	var op = "service.RemoveClient"

	err := p.Chats.Remove(chatID, client)
	if err != nil {
		p.Logger.Errorw("op", op, "err", ErrClientNotFound)

	}
}

func (p *PubSubService) validateMessage(msg *pb.Message) error {
	op := "service.validateMessage"

	chatID := msg.GetChatId()
	senderID := msg.GetSenderId()
	messageID := msg.GetMessageId()
	messageText := msg.GetMessageText()
	sentTime := msg.GetSentTs()

	if messageID == "" {
		return fmt.Errorf("%s: %w", op, ErrNoMessageID)
	}

	if chatID == "" || senderID == "" || messageText == "" || sentTime == "" {
		return fmt.Errorf("%s: error validating message %v: %w", op, messageID, ErrMessageMissingField)
	}

	return nil
}
