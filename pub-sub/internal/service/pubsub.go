package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/logging"
	"github.com/zoninnik89/messenger/pub-sub/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type PubSubService struct {
	Connections       *storage.ClientConnStorage
	ChatsParticipants *storage.ChatParticipantsStorage
	UserChats         *storage.UsersChats
	Logger            *zap.SugaredLogger
	chanBuffer        int
}

func NewPubSubService(chanBuffer int) *PubSubService {
	return &PubSubService{
		Connections:       storage.NewClientConnStorage(),
		ChatsParticipants: storage.NewChatParticipantsStorage(),
		UserChats:         storage.NewUsersChats(),
		Logger:            logging.GetLogger().Sugar(),
		chanBuffer:        chanBuffer,
	}
}

var (
	ErrChatNotExists       = errors.New("chat not exists")
	ErrNoChatSubscribers   = errors.New("no chat subscribers")
	ErrClientNotFound      = errors.New("client not found")
	ErrNoMessageID         = errors.New("no message ID")
	ErrMessageMissingField = errors.New("message misses one of the fields")
)

// Subscribe method used by chat client service to establish a stream for receiving messages
func (p *PubSubService) Subscribe(userID string, stream pb.PubSubService_SubscribeServer) error {
	var op = "service.Subscribe"

	p.Logger.Infow("start subscribing for user ID", "op", op, "user ID", userID)
	if userID == "" {
		return fmt.Errorf("%s: %s", "user ID required", userID)
	}

	channel := make(chan *pb.Message, p.chanBuffer)

	p.Connections.Add(userID, channel)

	// change for a call to DB or cache, now the data is taken just from a hashmap with few users and chats
	for _, chatID := range p.UserChats.Users[userID] {
		p.ChatsParticipants.Add(chatID, userID)
	}

	defer func() {
		p.Logger.Infow("removing client from connections", "userID", userID)
		p.removeClient(userID)
		close(channel)
	}()

	for {
		select {
		case msg := <-channel:
			p.Logger.Infow("start sending message to user channel", "op", op, "message", msg)
			if err := stream.Send(msg); err != nil {
				p.Logger.Errorw("error sending message to user", "op", op, "err", err)
				return fmt.Errorf("%s: %w", op, err)
			}
		case <-stream.Context().Done():
			p.Logger.Infow("user disconnected from server", "op", op, "user ID", userID)
			return nil
		}
	}
}

// ConsumeAndSendoutMessage method
func (p *PubSubService) ConsumeAndSendoutMessage(ctx context.Context, consumer *kafka.Consumer) (string, error) {
	var op = "service.ConsumeMessage"

	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		p.Logger.Fatalw("failed to read message", "op", op, "err", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var deserializedMessage pb.Message
	if err := proto.Unmarshal(msg.Value, &deserializedMessage); err != nil {
		p.Logger.Fatalw("failed to unmarshal message", "op", op, "err", err)
	}

	messageID := deserializedMessage.GetMessageId()

	p.Logger.Infow("validating message", "op", op, "message", messageID)

	if err := p.validateMessage(&deserializedMessage); err != nil {
		p.Logger.Errorw("invalid message", "op", op, "err", err)
		return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, err)
	}

	chatID := deserializedMessage.GetChatId()
	senderID := deserializedMessage.GetSenderId()
	messageText := deserializedMessage.GetMessageText()
	sentTime := deserializedMessage.GetSentTs()

	chatParticipants, err := p.ChatsParticipants.Get(chatID)
	if err != nil {
		return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, ErrChatNotExists)
	}

	// If there are no available recipients
	if len(chatParticipants.Store) <= 1 {
		return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, ErrNoChatSubscribers)
	}

	// Send the message to all clients
	for recipientID := range chatParticipants.Store {
		channel, err := p.Connections.Get(recipientID)
		if err != nil {
			p.Logger.Errorw("unsuccessful user chan retrieval", "op", op, "recipientID", recipientID, "error", err)
			continue
		}
		if senderID == recipientID {
			p.Logger.Errorw("sender ID == user ID, message not sent", "op", op, "recipientID", recipientID, "error", err)
			continue
		}
		p.Logger.Infow(
			"sending message to recipient",
			"op", op,
			"recipientID", recipientID,
			"chatID", chatID,
			"senderID", senderID,
			"messageID", messageID,
			"messageText", messageText,
			"sentTime", sentTime,
		)

		channel <- &pb.Message{
			ChatId:      chatID,
			SenderId:    senderID,
			MessageId:   messageID,
			MessageText: messageText,
			SentTs:      sentTime,
		}
	}
	p.Logger.Infow("message successfully sent out", "op", op, "chatID", chatID, "messageID", messageID)

	return messageID, nil
}

func (p *PubSubService) removeClient(userID string) {
	var op = "service.RemoveClient"

	err := p.Connections.Remove(userID)
	if err != nil {
		p.Logger.Errorw("unsuccessful user connection removal", "op", op, "err", ErrClientNotFound)
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
