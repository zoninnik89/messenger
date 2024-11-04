package grpc

import (
	"context"
	"errors"
	"github.com/zoninnik89/messenger/chat-history/internal/storage"
	"github.com/zoninnik89/messenger/chat-history/internal/types"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	pb.UnimplementedChatHistoryServiceServer
	service types.ChatHistoryServiceInterface
}

func Register(grpcServer *grpc.Server, s types.ChatHistoryServiceInterface) {
	pb.RegisterChatHistoryServiceServer(grpcServer, &serverAPI{service: s})
}

func (h *serverAPI) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	if err := h.validateGetMessagesRequest(ctx, req); err != nil {
		return nil, err
	}

	chatID := req.GetChatId()
	fromTS := req.GetFromTs()
	toTS := req.GetToTs()

	retrievedMessages, err := h.service.GetChatMessages(ctx, chatID, fromTS, toTS)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	response := &pb.GetChatMessagesResponse{
		Messages: make([]*pb.Message, 0, len(retrievedMessages)),
	}

	for _, msg := range retrievedMessages {
		response.Messages = append(response.Messages, &pb.Message{
			ChatId:      msg.ChatID,
			SenderId:    msg.SenderID,
			MessageId:   msg.ID,
			MessageText: msg.Text,
			SentTs:      msg.SentAt,
		})
	}

	return response, nil
}

func (h *serverAPI) GetChatsByUserID(ctx context.Context, req *pb.GetChatsListByParticipantIDRequest) (*pb.GetChatsListByParticipantIDResponse, error) {
	if err := h.validateGetChatsRequest(req); err != nil {
		return nil, err
	}

	userID := req.GetParticipantId()
	chats, err := h.service.GetChatsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	response := &pb.GetChatsListByParticipantIDResponse{
		Chats: make([]*pb.Chat, 0, len(chats)),
	}

	for _, chat := range chats {
		response.Chats = append(
			response.Chats,
			&pb.Chat{ID: chat.ID})
	}

	return response, nil
}

func (h *serverAPI) validateGetMessagesRequest(ctx context.Context, req *pb.GetChatMessagesRequest) error {
	if req.GetChatId() == "" {
		return status.Error(codes.InvalidArgument, "chat_id is required")
	}

	if _, err := h.service.GetChatByID(ctx, req.GetChatId()); err != nil {
		return status.Error(codes.NotFound, "chat not found")
	}

	if req.GetFromTs() < 0 {
		return status.Error(codes.InvalidArgument, "from_ts must be greater than zero")
	}

	if req.GetToTs() < 0 {
		return status.Error(codes.InvalidArgument, "to_ts must be greater than zero")
	}

	return nil
}

func (h *serverAPI) validateGetChatsRequest(req *pb.GetChatsListByParticipantIDRequest) error {
	if req.GetParticipantId() == "" {
		return status.Error(codes.InvalidArgument, "participant_id is required")
	}

	return nil
}
