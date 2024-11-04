package types

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "github.com/zoninnik89/messenger/common/api"
)

type ChatHistoryServiceInterface interface {
	ConsumeMessage(ctx context.Context, queue *kafka.Consumer) error
	GetMessages(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error)
	ConsumeMessageReadEvent(ctx context.Context, request *pb.SendMessageReadEventRequest) error
}

type StoreInterface interface {
	Add(ctx context.Context, chatID, senderID, messageID, messageText, sentTime string) error
	GetAll(ctx context.Context, chatID, fromTS, toTS string) ([]*pb.Message, error)
	AddReadEvent(ctx context.Context, chatId, messageId, readByUserId, readAt string) error
}
