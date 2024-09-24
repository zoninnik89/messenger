package auth

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
)

type serverAPI struct {
	pb.UnimplementedAuthServiceServer
}

func Register(s *grpc.Server) {
	pb.RegisterAuthServiceServer(s, &serverAPI{})
}

func (s *serverAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {

}

func (s *serverAPI) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	
}
