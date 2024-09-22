package service

import (
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/types"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	pb.UnimplementedPubSubServiceServer
	service types.PubSubServiceInterface
}

func NewGrpcHandler(grpcServer *grpc.Server, service types.PubSubServiceInterface) {
	handler := &GrpcHandler{service: service}
	pb.RegisterPubSubServiceServer(grpcServer, handler)
}

func (h *GrpcHandler) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	return h.service.Subscribe(req, stream)
}
