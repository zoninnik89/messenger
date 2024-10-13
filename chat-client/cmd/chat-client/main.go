package main

import (
	"context"
	"github.com/zoninnik89/messenger/chat-client/internal/app"
	"github.com/zoninnik89/messenger/chat-client/internal/config"
	"github.com/zoninnik89/messenger/chat-client/internal/logging"
	kafkaProducer "github.com/zoninnik89/messenger/chat-client/internal/producer"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	zap "go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.InitLogger()
	defer logging.Sync()

	logger.Info("starting chat-client service")

	registry, err := consul.NewRegistry(cfg.GRPC.Address, cfg.Consul.Port)
	if err != nil {
		logger.Panic("failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(cfg.GRPC.Name)
	if err := registry.Register(
		ctx,
		instanceID,
		cfg.GRPC.Address,
		cfg.GRPC.Port,
		cfg.GRPC.Name,
	); err != nil {

		logger.Panic("failed to register service", zap.Error(err))
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID); err != nil {
				logger.Warn("failed to health check", zap.Error(err))
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

	//ctxWithCancel, cancel := context.WithCancel(ctx)

	queue, err := kafkaProducer.NewKafkaProducer()
	if err != nil {
		logger.Panic("failed to connect to Kafka", zap.Error(err))
	}
	defer queue.Producer.Flush(10)

	application := app.NewApp(cfg.GRPC.Port, registry, queue)
	go application.GRPCsrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	s := <-stop

	logger.Info("shutting down gracefully", zap.Any("signal", s))

	//cancel()
	application.GRPCsrv.Stop()

	logger.Info("shut down gracefully")
}
