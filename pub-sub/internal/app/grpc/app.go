package grpcapp

import (
	"fmt"
	pubsubgrpc "github.com/zoninnik89/messenger/pub-sub/internal/grpc"
	"github.com/zoninnik89/messenger/pub-sub/internal/logging"
	"github.com/zoninnik89/messenger/pub-sub/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	logger     *zap.SugaredLogger
	grpcServer *grpc.Server
	service    *types.PubSubServiceInterface
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
		a.logger.Fatalw("failed to listen", "op", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	a.logger.Infow("grpc server is running", "add", l.Addr().String())

	if err := a.grpcServer.Serve(l); err != nil {
		a.logger.Fatalw("failed to serve", "op", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.logger.Infow("stopping grpc server")
	a.grpcServer.GracefulStop()
}
