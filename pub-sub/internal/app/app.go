package app

import (
	grpcapp "github.com/zoninnik89/messenger/pub-sub/internal/app/grpc"
	"github.com/zoninnik89/messenger/pub-sub/internal/service"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int, charBuffer int) *App {
	pubSubService := service.NewPubSubService(charBuffer)

	grpcApp := grpcapp.NewApp(pubSubService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
