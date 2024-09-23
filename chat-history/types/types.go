package types

import (
	pb "github.com/zoninnik89/messenger/common/api"
)

type ChatHistoryServiceInterface interface {
	GetMessages(request *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error)
	SendMessageReadEvent(request *pb.SendMessageReadEventRequest) (*pb.SendMessageReadEventResponse, error)
}
