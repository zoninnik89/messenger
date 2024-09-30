package app

import (
	grpcapp "github.com/zoninnik89/messenger/chat-client/internal/app/grpc"
	producer "github.com/zoninnik89/messenger/chat-client/internal/producer"
	"github.com/zoninnik89/messenger/chat-client/internal/service"
	"github.com/zoninnik89/messenger/common/discovery"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int, r discovery.Registry, queue *producer.MessageProducer) *App {
	chatClientService, err := service.NewChatClient(r, queue)
	if err != nil {
		return nil
	}

	grpcApp := grpcapp.NewApp(chatClientService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
