package grpcapp

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	chathistorygrpc "github.com/zoninnik89/messenger/chat-history/internal/grpc"
	"github.com/zoninnik89/messenger/chat-history/internal/logging"
	"github.com/zoninnik89/messenger/chat-history/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"time"
)

type App struct {
	logger     *zap.SugaredLogger
	grpcServer *grpc.Server
	service    types.ChatHistoryServiceInterface
	port       int
}

func NewApp(
	chatHistoryService types.ChatHistoryServiceInterface,
	port int,
) *App {

	l := logging.GetLogger().Sugar()
	grpcServer := grpc.NewServer()
	chathistorygrpc.Register(grpcServer, chatHistoryService)

	return &App{
		logger:     l,
		grpcServer: grpcServer,
		port:       port,
		service:    chatHistoryService,
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
			a.logger.Warnw("error consuming a message", "op", op, err, zap.Error(err))
		} else {
			a.logger.Infow("message was consumed", "op", op, "status", status)
		}

		time.Sleep(time.Second * 1)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	a.logger.Infow("starting grpc app", "op", op, "port", a.port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		a.logger.Fatalw(op, "failed to listen", "op", op, "port", a.port)
	}

	addr := lis.Addr().String()

	a.logger.Infow("grpc server is listening", "op", op, "port", a.port)

	if err := a.grpcServer.Serve(lis); err != nil {
		a.logger.Fatalw(op, "failed to serve", "op", op, "port", a.port)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.logger.Infow("stopping grpc app", "op", op, "port", a.port)
	a.grpcServer.GracefulStop()
}
