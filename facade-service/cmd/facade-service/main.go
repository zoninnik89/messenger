package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	"github.com/zoninnik89/messenger/facade-service/internal/config"
	grpcgateway "github.com/zoninnik89/messenger/facade-service/internal/gateway"
	"github.com/zoninnik89/messenger/facade-service/internal/http-server/handlers/auth/login"
	"github.com/zoninnik89/messenger/facade-service/internal/http-server/handlers/auth/register"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	websocketserver "github.com/zoninnik89/messenger/facade-service/internal/websocket-server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.InitLogger()
	defer logging.Sync()

	strPort := strconv.Itoa(cfg.HTTPServer.Port)

	logger.Info("starting facade service", zap.String("port", strPort))

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

	gateway := grpcgateway.NewGRPCGateway(registry)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// Basic CORS configuration
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},                                       // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Allow specific methods
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // Allow cookies to be sent
		MaxAge:           300,  // Maximum value for the preflight request cache
	}))

	router.Post("/login", login.New(gateway))
	router.Post("/register", register.New(gateway))

	wsServer := websocketserver.NewWebsocketServer(gateway)

	router.Get("/ws", wsServer.ServeHTTP)

	// Start the server using http.Serve with the custom listener
	logger.Info("http server is listening", zap.String("port", "3002"))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", "3002"), router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

}
