package types

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
)

type ChatHistoryServiceInterface interface {
	StoreMessage(ctx context.Context, message *pb.Message) error
	GetMessages(ctx context.Context, request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error)
	StoreMessageReadEvent(ctx context.Context, request *pb.SendMessageReadEventRequest) error
}

type StoreInterface interface {
	Add(ctx context.Context, chatID, senderID, messageID, messageText, sentTime string) error
	GetAll(ctx context.Context, chatID, fromTS, toTS string) ([]*pb.Message, error)
	AddReadEvent(ctx context.Context, chatId, messageId, readByUserId, readAt string) error
}
