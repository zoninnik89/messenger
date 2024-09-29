package app

import (
	grpcapp "github.com/zoninnik89/messenger/chat-client/internal/app/grpc"
	"github.com/zoninnik89/messenger/chat-client/internal/service"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int) *App {
	chatClientService := service.NewChatClient()

	grpcApp := grpcapp.NewApp(chatClientService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
