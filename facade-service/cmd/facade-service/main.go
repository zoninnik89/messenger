package main

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	"github.com/zoninnik89/messenger/facade-service/internal/config"
	"github.com/zoninnik89/messenger/facade-service/internal/http-server/handlers/chat/send-message"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	"go.uber.org/zap"
	"time"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.InitLogger()
	defer logging.Sync()

	logger.Info("starting facade service")

	registry, err := consul.NewRegistry(cfg.HTTPServer.Address, cfg.Consul.Port)
	if err != nil {
		logger.Panic("failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(cfg.HTTPServer.Name)
	if err := registry.Register(
		ctx,
		instanceID,
		cfg.HTTPServer.Address,
		cfg.HTTPServer.Port,
		cfg.HTTPServer.Name,
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

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/send", send_message.New())
}
