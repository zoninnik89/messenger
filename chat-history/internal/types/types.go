package types

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/chat-history/internal/domain/models"
)

type ChatHistoryServiceInterface interface {
	ConsumeMessage(ctx context.Context, queue *kafka.Consumer) error
	GetChatMessages(ctx context.Context, chatID string, fromTS, toTS int64) ([]models.Message, error)
	GetChatsByUserID(ctx context.Context, userID string) ([]models.Chat, error)
	GetChatByID(ctx context.Context, chatID string) (models.Chat, error)
}

type StoreInterface interface {
	SaveMessage(ctx context.Context, messageID string, chatID string, senderID string, messageText string, sentTS int64) error
	GetChatMessages(ctx context.Context, chatID string, fromTS int64, toTS int64) ([]models.Message, error)
	GetChatsByUserID(ctx context.Context, userID string) ([]models.Chat, error)
	GetChatByID(ctx context.Context, chatID string) (models.Chat, error)
}
