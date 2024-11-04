package app

import (
	grpcapp "github.com/zoninnik89/messenger/chat-history/internal/app/grpc"
	"github.com/zoninnik89/messenger/chat-history/internal/service"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int, storagePath string) *App {
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}
	chatHistoryService := service.NewChatHistoryService(storage)
	grpcApp := grpcapp.NewApp(chatHistoryService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
