package main

import (
	"context"
	"fmt"
	c "github.com/zoninnik89/messenger/chat-history/internal/consumer"
	h "github.com/zoninnik89/messenger/chat-history/internal/grpc"
	"github.com/zoninnik89/messenger/chat-history/internal/logging"
	s "github.com/zoninnik89/messenger/chat-history/internal/service"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	zap "go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc"
	"net"
	_ "strconv"
	"time"
)

func main() {
	logger := logging.InitLogger()
	defer logging.Sync()

	registry, err := consul.NewRegistry(consulAddress, serviceName)
	if err != nil {
		logger.Panic("Failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddress); err != nil {
		logger.Panic("Failed to register service", zap.Error(err))
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logger.Warn("Failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer func(registry *consul.Registry, ctx context.Context, instanceID string, serviceName string) {
		err := registry.Deregister(ctx, instanceID, serviceName)
		if err != nil {
			logger.Fatal("Failed to deregister service", zap.Error(err))
		}
	}(registry, ctx, instanceID, serviceName)

	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		logger.Fatal("Failed to listen:", zap.Error(err))
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			logger.Warn("Failed to close listener", zap.Error(err))
		}
	}(l)

	uri := fmt.Sprintf("mongodb://%s:%s@%s", mongoUser, mongoPass, mongoAddr)
	mongoClient, err := connectToMongoDB(uri)
	if err != nil {
		logger.Fatal("Failed to connect to mongodb", zap.Error(err))
	}

	store := NewStore(mongoClient)

	service := s.NewChatHistoryService(store)
	h.NewGrpcHandler(grpcServer, service)

	logger.Info("Starting GRPC server", zap.String("port", grpcAddress))

	logger.Info("Starting Kafka Consumer")
	consumer, err := c.NewKafkaConsumer()
	if err != nil {
		logger.Panic("Failed to create kafka consumer", zap.Error(err))
		panic(err)
	}

	topics := []string{"messages", "read_events"}
	err = consumer.SubscribeTopics(topics, nil)

	if err != nil {
		panic(err)
	}

	go func() {
		m, err := service.ConsumeMessage(ctx, consumer)
		if err != nil {
			logger.Fatal("Error consuming a message", zap.Error(err))
		} else {
			logger.Info("Message was consumed with status", zap.String("message", m.Status))
		}

		time.Sleep(time.Second * 1)
	}()

	if err := grpcServer.Serve(l); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
