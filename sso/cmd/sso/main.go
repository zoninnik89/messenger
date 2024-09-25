package main

import (
	"github.com/zoninnik89/messenger/sso/internal/app"
	"github.com/zoninnik89/messenger/sso/internal/config"
	"github.com/zoninnik89/messenger/sso/internal/logging"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.InitLogger().Sugar()

	logger.Info("Starting sso service")

	application := app.NewApp(cfg.GRPC.Port, logger, cfg.StoragePath, cfg.TokenTTL)
	go application.GRPCsrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	s := <-stop

	logger.Infow("Shutting down gracefully", "signal", s)

	application.GRPCsrv.Stop()

	logger.Info("Shut down gracefully")
}
