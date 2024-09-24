package service

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/chat-history/types"
	pb "github.com/zoninnik89/messenger/common/api"
	"go.uber.org/zap"
	"strings"
)

type ChatHistoryService struct {
	store  types.StoreInterface
	logger *zap.SugaredLogger
}

func NewChatHistoryService(s types.StoreInterface) *ChatHistoryService {
	return &ChatHistoryService{store: s}
}

func (s *ChatHistoryService) StoreMessage(ctx context.Context, consumer *kafka.Consumer) error {

	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		s.logger.Fatalw("Failed to read message", "err", err)
		return err
	}
	msgSlice := strings.Split(string(msg.Value), ",")
	chatID, senderID, messageID, messageText, sentTime := msgSlice[0], msgSlice[1], msgSlice[2], msgSlice[3], msgSlice[4]

	err = s.store.Add(ctx, chatID, senderID, messageID, messageText, sentTime)
	if err != nil {
		return err
	}

	return nil
}

func (s *ChatHistoryService) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	messages, err := s.store.GetAll(ctx, req.ChatId, req.FromTs, req.ToTs)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	return &pb.GetMessagesResponse{Message: messages}, nil
}

func (s *ChatHistoryService) StoreMessageReadEvent(ctx context.Context, req *pb.SendMessageReadEventRequest) error {
	err := s.store.AddReadEvent(ctx, req.ChatId, req.MessageId, req.ReadByUserId, req.ReadAt)
	if err != nil {
		s.logger.Error(err)
		return err
	}
	return nil
}
