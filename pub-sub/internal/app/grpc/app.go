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
		status, err := a.service.ConsumeAndSendoutMessage(ctx, consumer)
		if err != nil {
			a.logger.Warnw("error consuming a message", "op", op, err, zap.Error(err))
		} else {
			a.logger.Infow("message was consumed", "op", op, "status", status)
		}

		time.Sleep(time.Second * 1)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	a.logger.Infow("starting grpc server", "op", op, "port", a.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		a.logger.Fatalw("failed to listen", "op", op, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	addr := l.Addr().String()

	a.logger.Infow("grpc server is listening", "op", op, "addr", addr)

	if err := a.grpcServer.Serve(l); err != nil {
		a.logger.Fatalw("failed to serve", "op", op, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.logger.Infow("stopping grpc server", "op", op, "port", a.port)
	a.grpcServer.GracefulStop()
}
