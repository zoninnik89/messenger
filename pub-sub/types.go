package main

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
)

type PubSubServiceInterface interface {
	Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error
	Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error)
}
