package grpcapp

import (
	"fmt"
	authgrpc "github.com/zoninnik89/messenger/sso/internal/grpc/auth"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"net"
)

type App struct {
	logger     *zap.SugaredLogger
	grpcServer *grpc.Server
	port       int
}

func NewApp(l *zap.SugaredLogger, port int) *App {
	grpcServer := grpc.NewServer()
	authgrpc.Register(grpcServer)

	return &App{grpcServer: grpcServer, logger: l, port: port}
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
