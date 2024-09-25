package app

import (
	grpcapp "github.com/zoninnik89/messenger/sso/internal/app/grpc"
	"github.com/zoninnik89/messenger/sso/internal/services/auth"
	"github.com/zoninnik89/messenger/sso/internal/storage/sqlite"
	"go.uber.org/zap"
	"time"
)

type App struct {
	GRPCsrv *grpcapp.App
}

func NewApp(grpcPort int, logger *zap.SugaredLogger, storagePath string, tokenTTL time.Duration) *App {
	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.NewAuthService(logger, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.NewApp(logger, authService, grpcPort)

	return &App{
		GRPCsrv: grpcApp,
	}
}
