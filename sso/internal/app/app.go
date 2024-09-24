package app

import (
	grpcapp "github.com/zoninnik89/messenger/sso/internal/app/grpc"
	"go.uber.org/zap"
	"time"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int, logger *zap.SugaredLogger, storagePath string, tokenTTL time.Duration) *App {
	grpcApp := grpcapp.NewApp(logger, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
