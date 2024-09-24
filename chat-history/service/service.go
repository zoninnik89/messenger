package service

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/chat-history/logging"
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
	l := logging.GetLogger().Sugar()
	return &ChatHistoryService{store: s, logger: l}
}

func (s *ChatHistoryService) ConsumeMessage(ctx context.Context, queue *kafka.Consumer) (*pb.Message, error) {
	msg, err := queue.ReadMessage(-1)
	if err != nil {
		s.logger.Fatalw("Failed to read message", "err", err)
		return nil, err
	}
	msgSlice := strings.Split(string(msg.Value), ",")
	chatID, senderID, messageID, messageText, sentTime := msgSlice[0], msgSlice[1], msgSlice[2], msgSlice[3], msgSlice[4]

	err = s.store.Add(ctx, chatID, senderID, messageID, messageText, sentTime)
	if err != nil {
		return nil, err
	}

	return &pb.Message{ChatId: chatID,
		MessageId:   messageID,
		MessageText: messageText,
		SenderId:    senderID,
		SentTs:      sentTime}, nil
}

func (s *ChatHistoryService) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	messages, err := s.store.GetAll(ctx, req.ChatId, req.FromTs, req.ToTs)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	return &pb.GetMessagesResponse{Message: messages}, nil
}

func (s *ChatHistoryService) ConsumeMessageReadEvent(ctx context.Context, req *pb.SendMessageReadEventRequest) error {
	err := s.store.AddReadEvent(ctx, req.ChatId, req.MessageId, req.ReadByUserId, req.ReadAt)
	if err != nil {
		s.logger.Error(err)
		return err
	}
	return nil
}
