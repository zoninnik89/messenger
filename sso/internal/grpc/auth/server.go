package auth

import (
	"context"
	"errors"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/sso/internal/services/auth"
	"github.com/zoninnik89/messenger/sso/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedAuthServiceServer
	service types.Auth
}

func Register(srv *grpc.Server, svs types.Auth) {
	pb.RegisterAuthServiceServer(srv, &serverAPI{service: svs})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := validateLoginData(req); err != nil {
		return nil, err
	}

	token, err := s.service.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		if errors.Is(err, auth.ErrInvalidAppID) {
			return nil, status.Error(codes.InvalidArgument, "invalid app id")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateRegisterData(req); err != nil {
		return nil, err
	}

	userID, err := s.service.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.RegisterResponse{
		UserId: userID,
	}, nil
}

func validateLoginData(req *pb.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password required")
	}
	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app id required")
	}

	return nil
}

func validateRegisterData(req *pb.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password required")
	}

	return nil
}
