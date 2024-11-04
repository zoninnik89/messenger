package service

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/chat-history/internal/domain/models"
	"github.com/zoninnik89/messenger/chat-history/internal/logging"
	"github.com/zoninnik89/messenger/chat-history/internal/types"
	"go.uber.org/zap"
	"strconv"
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

func (s *ChatHistoryService) ConsumeMessage(ctx context.Context, queue *kafka.Consumer) error {
	msg, err := queue.ReadMessage(-1)
	if err != nil {
		return err
	}
	msgSlice := strings.Split(string(msg.Value), ",")
	chatID, senderID, messageID, messageText, sentTime := msgSlice[0], msgSlice[1], msgSlice[2], msgSlice[3], msgSlice[4]

	sentTimeConverted, err := strconv.ParseInt(sentTime, 10, 64)
	if err != nil {
		return err
	}

	err = s.store.SaveMessage(ctx, messageID, chatID, senderID, messageText, sentTimeConverted)
	if err != nil {
		return err
	}

	return nil
}

func (s *ChatHistoryService) GetChatMessages(ctx context.Context, chatID string, fromTS, toTS int64) ([]models.Message, error) {
	messages, err := s.store.GetChatMessages(ctx, chatID, fromTS, toTS)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	return messages, nil
}

func (s *ChatHistoryService) GetChatsByUserID(ctx context.Context, userID string) ([]models.Chat, error) {
	chats, err := s.store.GetChatsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error(err)
		return nil, err
	}
	return chats, nil
}

func (s *ChatHistoryService) GetChatByID(ctx context.Context, chatID string) (models.Chat, error) {
	chat, err := s.store.GetChatByID(ctx, chatID)

	return chat, err
}
