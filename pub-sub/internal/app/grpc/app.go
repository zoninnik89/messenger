package grpcapp

import (
	"context"
	"fmt"
	c "github.com/zoninnik89/messenger/chat-history/consumer"
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
	}
}

func (a *App) MustRun(ctx context.Context, kafkaPort string, kafkaConsumerID string, kafkaConsumerGroup string) {

	if err := a.Run(); err != nil {
		panic(err)
	}

	a.logger.Info("Starting Kafka Consumer")
	consumer, err := c.NewKafkaConsumer(kafkaPort, kafkaConsumerID, kafkaConsumerGroup)
	if err != nil {
		a.logger.Panic("failed to create kafka consumer", zap.Error(err))
		panic(err)
	}

	topics := []string{"messages", "read_events"}
	err = consumer.SubscribeTopics(topics, nil)

	if err != nil {
		panic(err)
	}

	go func() {
		status, err := a.service.ConsumeMessage(ctx, consumer)
		if err != nil {
			a.logger.Warn("error consuming a message", zap.Error(err))
		} else {
			a.logger.Info("Message was consumed with status", zap.String("message", status))
		}

		time.Sleep(time.Second * 1)
	}()
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
