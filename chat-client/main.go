package main

import (
	"context"
	common "github.com/zoninnik89/messenger/common"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/common/discovery/consul"
	zap "go.uber.org/zap"
	"log"
	"time"
)

var (
	serviceName   = "chat-client"
	grpcAddress   = common.EnvString("GRPC_ADDR", ":2001")
	consulAddress = common.EnvString("CONSUL_ADDR", ":8500")
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	registry, err := consul.NewRegistry(consulAddress, serviceName)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddress); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatal("failed to health check")
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	newGateway := discovery.Registry(registry)

	conn, err := discovery.ServiceConnection(ctx, "pub-sub", newGateway)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

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
