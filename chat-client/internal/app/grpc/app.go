package grpcapp

import (
	"fmt"
	chatclientgrpc "github.com/zoninnik89/messenger/chat-client/internal/grpc"
	"github.com/zoninnik89/messenger/chat-client/internal/logging"
	"github.com/zoninnik89/messenger/chat-client/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	logger     *zap.SugaredLogger
	grpcServer *grpc.Server
	service    types.ChatClientInterface
	port       int
}

func NewApp(chatClientService types.ChatClientInterface, port int) *App {
	l := logging.GetLogger().Sugar()
	grpcServer := grpc.NewServer()
	chatclientgrpc.Register(grpcServer, chatClientService)

	return &App{
		logger:     l,
		grpcServer: grpcServer,
		port:       port,
		service:    chatClientService,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	a.logger.Infow("starting grpc server", "op", op, "port", a.port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
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
