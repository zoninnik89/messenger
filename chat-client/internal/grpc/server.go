package grpc

import (
	"context"
	"github.com/zoninnik89/messenger/chat-client/internal/types"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedChatClientServiceServer
	service types.ChatClientInterface
}

func Register(srv *grpc.Server, service types.ChatClientInterface) {
	pb.RegisterChatClientServiceServer(srv, &serverAPI{service: service})
}

func (s *serverAPI) GetMessagesStream(req *pb.GetMessagesStreamRequest, stream pb.ChatClientService_GetMessagesStreamServer) error {
	const op = "gateway.SubscribeForMessages"
	userID := req.GetUserId()

	if err := validateUser(userID); err != nil {
		return err
	}

	err := s.service.SubscribeForMessages(context.Background(), userID, stream)
	if err != nil {
		return status.Error(codes.Internal, "internal server error")
	}

	return nil
}

func (s *serverAPI) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	const op = "gateway.SendMessage"

	if err := validateMessage(req); err != nil {
		return nil, err
	}

	err := s.service.SendMessage(
		req.Message.GetMessageId(),
		req.Message.GetChatId(),
		req.Message.GetSenderId(),
		req.Message.GetMessageText(),
		req.Message.GetSentTs(),
	)

	if err != nil {
		// add error handling
	}

	return &pb.SendMessageResponse{Status: "sent"}, nil
}

func validateUser(userID string) error {
	if userID == "" {
		return status.Errorf(codes.InvalidArgument, "user id is required")
	}
	return nil
}

func validateMessage(req *pb.SendMessageRequest) error {
	const op = "gateway.validateMessage"

	chatID := req.Message.GetChatId()
	senderID := req.Message.GetSenderId()
	messageID := req.Message.GetMessageId()
	messageText := req.Message.GetMessageText()
	sentTime := req.Message.GetSentTs()

	if messageID == "" {
		return status.Error(codes.InvalidArgument, "message ID is required")
	}

	if chatID == "" {
		return status.Error(codes.InvalidArgument, "chat ID is required")
	}

	if senderID == "" {
		return status.Error(codes.InvalidArgument, "sender ID is required")
	}

	if messageText == "" {
		return status.Error(codes.InvalidArgument, "message text is required")
	}

	if sentTime == "" {
		return status.Error(codes.InvalidArgument, "sent timestamp is required")
	}

	return nil
}
