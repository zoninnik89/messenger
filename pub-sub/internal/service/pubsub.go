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
	"strconv"
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
	for i := 1; i <= 5; i++ {
		p.ChatsParticipants.Add(strconv.Itoa(i), userID)
		p.Logger.Infof("subscribed user: %s to chat: %s", userID, strconv.Itoa(i))
	}

	defer func() {
		p.Logger.Infow("removing user connection from connections storage", "userID", userID)
		p.removeUserConnection(userID)
		close(channel)
	}()

	p.Logger.Infow("User subscribed for messages", "op", op, "user ID", userID)

	for {
		select {
		case msg := <-channel:
			p.Logger.Infow("sending message", "op", op, "recipient ID", userID, "message", msg)
			if err := stream.Send(msg); err != nil {
				p.Logger.Errorw("error sending message to user", "op", op, "err", err)
				return fmt.Errorf("%s: %w", op, err)
			}
			p.Logger.Infow("message sent", "op", op, "recipient ID", userID, "message", msg)
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
	//if len(chatParticipants.Store) <= 1 {
	//	return "", fmt.Errorf("%s: error validating message %v: %w", op, messageID, ErrNoChatSubscribers)
	//}

	// Send the message to all clients
	for recipientID := range chatParticipants.Store {
		channel, err := p.Connections.Get(recipientID)
		if err != nil {
			p.Logger.Errorw("unsuccessful user chan retrieval", "op", op, "recipientID", recipientID, "error", err)
		} else {
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
	}
	p.Logger.Infow("message successfully sent out", "op", op, "chatID", chatID, "messageID", messageID)

	return messageID, nil
}

func (p *PubSubService) removeUserConnection(userID string) {
	var op = "service.RemoveClient"

	err := p.Connections.Remove(userID)
	if err != nil {
		p.Logger.Errorw("unsuccessful user connection removal", "op", op, "err", ErrClientNotFound)
		return
	}
	p.Logger.Infow("user connection removed", "op", op, "user", userID)
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
