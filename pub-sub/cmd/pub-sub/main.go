package main

import (
	"context"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	"github.com/zoninnik89/messenger/pub-sub/internal/app"
	"github.com/zoninnik89/messenger/pub-sub/internal/config"
	"github.com/zoninnik89/messenger/pub-sub/internal/logging"
	zap "go.uber.org/zap"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.InitLogger()
	defer logging.Sync()

	logger.Info("starting sso service")

	registry, err := consul.NewRegistry(cfg.Env, cfg.Consul.Port)
	if err != nil {
		logger.Panic("failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(cfg.GRPC.Name)
	if err := registry.Register(ctx, instanceID, cfg.Env, cfg.GRPC.Name, cfg.GRPC.Port); err != nil {
		logger.Panic("failed to register service", zap.Error(err))
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID); err != nil {
				logger.Warn("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer func(registry *consul.Registry, ctx context.Context, instanceID string) {
		err := registry.Deregister(ctx, instanceID)
		if err != nil {
			logger.Fatal("failed to deregister service", zap.Error(err))
		}
	}(registry, ctx, instanceID)

	ctxWithCancel, cancel := context.WithCancel(ctx)

	application := app.NewApp(cfg.GRPC.Port, cfg.Storage.ChanBuffer)
	go application.GRPCsrv.MustRun(ctxWithCancel, strconv.Itoa(cfg.Kafka.Port), cfg.Kafka.ConsumerID, cfg.Kafka.ConsumerGroup)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	s := <-stop

	logger.Info("shutting down gracefully", zap.Any("signal", s))

	cancel()
	application.GRPCsrv.Stop()

	logger.Info("shut down gracefully")
}
