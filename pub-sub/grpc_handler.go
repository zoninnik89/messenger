package main

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	pb.UnimplementedPubSubServiceServer
	service PubSubServiceInterface
}

func NewGrpcHandler(grpcServer *grpc.Server, service PubSubServiceInterface) {
	handler := &GrpcHandler{service: service}
	pb.RegisterPubSubServiceServer(grpcServer, handler)
}

func (h *GrpcHandler) Publish(ctx context.Context, request *pb.PublishRequest) (*pb.PublishResponse, error) {
	return h.service.Publish(ctx, request)
}

func (h *GrpcHandler) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	return h.service.Subscribe(req, stream)
}
