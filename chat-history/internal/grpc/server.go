package grpc

import (
	"context"
	"github.com/zoninnik89/messenger/chat-history/internal/types"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedChatHistoryServiceServer
	service types.ChatHistoryServiceInterface
}

func Register(grpcServer *grpc.Server, s types.ChatHistoryServiceInterface) {
	pb.RegisterChatHistoryServiceServer(grpcServer, &serverAPI{service: s})
}

func (h *serverAPI) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	res, err := h.service.GetMessages(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return res, nil
}

func validateRequest(req *pb.GetMessagesRequest) error {
	if req.GetChatId() == "" {
		return status.Error(codes.InvalidArgument, "chat_id is required")
	}

	// Add check that chat exists in the database

	return nil
}
