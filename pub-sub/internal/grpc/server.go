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
	if err := validateUser(req); err != nil {
		return err
	}

	err := h.service.Subscribe(req.GetUserId(), stream)
	if err != nil {
		return status.Error(codes.Internal, "internal server error")
	}

	return nil
}

func validateUser(req *pb.SubscribeRequest) error {
	if req.GetUserId() == "" {
		return status.Errorf(codes.InvalidArgument, "user id is required")
	}
	return nil
}
