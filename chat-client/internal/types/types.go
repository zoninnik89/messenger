package types

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
)

type ChatClientInterface interface {
	SubscribeForMessages(ctx context.Context, userID string, stream pb.ChatClientService_GetMessagesStreamServer) error
	SendMessage(messageID string, chatID string, senderID string, messageText string, sentTime string) error
}
