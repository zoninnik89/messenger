package grpc

import (
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedPubSubServiceServer
	service types.PubSubServiceInterface
}

func Register(srv *grpc.Server, service types.PubSubServiceInterface) {
	pb.RegisterPubSubServiceServer(srv, &serverAPI{service: service})
}

func (h *serverAPI) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	if err := validateChat(req); err != nil {
		return err
	}

	err := h.service.Subscribe(req.GetChatId(), stream)
	if err != nil {
		return status.Error(codes.Internal, "internal server error")
	}

	return nil
}

func validateChat(req *pb.SubscribeRequest) error {
	if req.GetChatId() == "" {
		return status.Errorf(codes.InvalidArgument, "chat ID required")
	}
	return nil
}
