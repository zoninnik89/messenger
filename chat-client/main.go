package main

import (
	"context"
	"github.com/zoninnik89/messenger/chat-client/logging"
	common "github.com/zoninnik89/messenger/common"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	zap "go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

var (
	serviceName   = "chat-client"
	grpcAddress   = common.EnvString("GRPC_ADDR", ":2001")
	consulAddress = common.EnvString("CONSUL_ADDR", ":8500")
)

func main() {
	logger := logging.InitLogger()
	defer logging.Sync()

	registry, err := consul.NewRegistry(consulAddress, serviceName)
	if err != nil {
		logger.Panic("failed to connect to Consul", zap.Error(err))
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddress); err != nil {
		logger.Panic("failed to register service", zap.Error(err))
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				logger.Warn("failed to health check", zap.Error(err))
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer func(registry *consul.Registry, ctx context.Context, instanceID string, serviceName string) {
		err := registry.Deregister(ctx, instanceID, serviceName)
		if err != nil {
			logger.Fatal("failed to deregister service", zap.Error(err))
		}
	}(registry, ctx, instanceID, serviceName)

	newGateway := discovery.Registry(registry)

	conn, err := discovery.ServiceConnection(ctx, "pub-sub", newGateway)
	if err != nil {
		logger.Fatal("failed to dial pub-sub server", zap.Error(err))
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			logger.Warn("failed to close connection with Pub-Sub server", zap.Error(err))
		}
	}(conn)

	client := NewChatClient(conn)

	go client.SubscribeToChat("chatroom1")

	time.Sleep(2 * time.Second)

	client.SendMessage("chatroom1", "Hello, everyone!")

	select {}

	//grpcServer := grpc.NewServer()
	//
	//listner, err := net.Listen("tcp", grpcAddress)
	//if err != nil {
	//	logger.Fatal("failed to listen:", zap.Error(err))
	//}
	//defer listner.Close()

	//service := NewChatClient()
	//NewGrpcHandler(grpcServer, service)
	//
	//logger.Info("Starting HTTP server", zap.String("port", grpcAddress))
	//
	//if err := grpcServer.Serve(listner); err != nil {
	//	logger.Fatal("failed to serve", zap.Error(err))
	//}

}
