package handlers

import (
	"context"
	"github.com/zoninnik89/messenger/chat-history/logging"
	"github.com/zoninnik89/messenger/chat-history/types"
	pb "github.com/zoninnik89/messenger/common/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	pb.UnimplementedChatHistoryServiceServer
	logger  *zap.SugaredLogger
	service types.ChatHistoryServiceInterface
}

func NewGrpcHandler(grpcServer *grpc.Server, s types.ChatHistoryServiceInterface) {
	l := logging.GetLogger().Sugar()
	handler := &GrpcHandler{service: s, logger: l}
	pb.RegisterChatHistoryServiceServer(grpcServer, handler)
}

func (h *GrpcHandler) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	res, err := h.service.GetMessages(ctx, req)
	if err != nil {
		h.logger.Errorw("error getting messages", "chatID", req.ChatId, "fromTS", req.FromTs, "toTS", req.ToTs, "error", err)
	}

	h.logger.Infow("Successfully retrieved messages", "chatID", req.ChatId, "fromTS", req.FromTs, "toTS", req.ToTs)
	return res, nil
}
