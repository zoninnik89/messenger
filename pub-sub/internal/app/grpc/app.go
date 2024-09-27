package grpcapp

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pubsubgrpc "github.com/zoninnik89/messenger/pub-sub/internal/grpc"
	"github.com/zoninnik89/messenger/pub-sub/internal/logging"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"time"
)

type App struct {
	logger     *zap.SugaredLogger
	grpcServer *grpc.Server
	service    types.PubSubServiceInterface
	port       int
}

func NewApp(
	pubSubService types.PubSubServiceInterface,
	port int) *App {

	l := logging.GetLogger().Sugar()
	grpcServer := grpc.NewServer()
	pubsubgrpc.Register(grpcServer, pubSubService)

	return &App{
		logger:     l,
		grpcServer: grpcServer,
		port:       port,
		service:    pubSubService,
	}
}

func (a *App) MustRun() {

	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) MustConsume(ctx context.Context, consumer *kafka.Consumer) {
	const op = "grpcapp.MustConsume"

	for {
		status, err := a.service.ConsumeMessage(ctx, consumer)
		if err != nil {
			a.logger.Warn("op", op, "error consuming a message", zap.Error(err))
		} else {
			a.logger.Info("op", op, "Message was consumed with status", zap.String("message", status))
		}

		time.Sleep(time.Second * 1)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	a.logger.Infow("op", op, "starting grpc server", "port", a.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		a.logger.Fatalw("op", op, "failed to listen", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	a.logger.Infow("op", op, "grpc server is running", l.Addr().String())

	if err := a.grpcServer.Serve(l); err != nil {
		a.logger.Fatalw("op", op, "failed to serve", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.logger.Infow("op", op, "stopping grpc server")
	a.grpcServer.GracefulStop()
}
