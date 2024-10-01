package grpcgateway

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
)

type ChatGateway interface {
	SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
	GetMessagesStream(req *pb.GetMessagesStreamRequest, stream pb.ChatClientService_GetMessagesStreamClient) error
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error)
}
