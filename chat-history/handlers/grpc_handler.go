package handlers

import (
	"context"
	"github.com/zoninnik89/messenger/chat-history/types"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	pb.UnimplementedChatHistoryServiceServer
	service types.ChatHistoryServiceInterface
}

func NewGrpcHandler(grpcServer *grpc.Server, s types.ChatHistoryServiceInterface) {
	handler := &GrpcHandler{service: s}
	pb.RegisterChatHistoryServiceServer(grpcServer, handler)
}

func (h *GrpcHandler) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	return h.service.GetMessages(req)
}

func (h *GrpcHandler) SendMessageReadEvent(ctx context.Context, req *pb.SendMessageReadEventRequest) (*pb.SendMessageReadEventResponse, error) {
	return h.service.SendMessageReadEvent(req)
}
