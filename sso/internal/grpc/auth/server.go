package auth

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedAuthServiceServer
}

func Register(s *grpc.Server) {
	pb.RegisterAuthServiceServer(s, &serverAPI{})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password required")
	}
	if req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "app id required")
	}

	return &pb.LoginResponse{
		Token: "",
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	panic("implement me")
}
