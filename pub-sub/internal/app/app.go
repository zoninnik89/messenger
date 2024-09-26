package app

import (
	grpcapp "github.com/zoninnik89/messenger/pub-sub/internal/app/grpc"
	"github.com/zoninnik89/messenger/pub-sub/internal/service"
	"time"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int, tokenTTL time.Duration) *App {
	pubSubService := service.NewPubSubService()

	grpcApp := grpcapp.NewApp(pubSubService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
