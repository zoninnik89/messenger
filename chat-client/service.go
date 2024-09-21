package main

import (
	"context"
	"github.com/zoninnik89/messenger/chat-client/logging"
	pb "github.com/zoninnik89/messenger/common/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
)

type ChatClient struct {
	client pb.PubSubServiceClient
	logger *zap.SugaredLogger
}

func NewChatClient(conn *grpc.ClientConn) *ChatClient {
	return &ChatClient{
		client: pb.NewPubSubServiceClient(conn),
		logger: logging.GetLogger().Sugar(),
	}
}

func (c *ChatClient) SubscribeToChat(chat string) {
	stream, err := c.client.Subscribe(context.Background(), &pb.SubscribeRequest{Chat: chat})
	if err != nil {
		c.logger.Fatalw("failed to subscribe to chat", "err", err)
		return
	}

	c.logger.Infow("Subscribed to chat", "chat", chat)

	// Continuously receive messages from the stream
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			// Server closed connection
			break
		}
		if err != nil {
			c.logger.Fatalw("failed to receive a message", "err", err)
		}

		// Log the received message
		c.logger.Infow("Received a message", "message", msg.Message, "chat", chat)
	}
}

func (c *ChatClient) SendMessage(chat, message string) {
	// publish message to the chat
	resp, err := c.client.Publish(context.Background(), &pb.PublishRequest{
		Chat:    chat,
		Message: message,
	})
	if err != nil {
		c.logger.Fatalw("failed to send the message", "err", err)
		return
	}

	c.logger.Infow("Message sent", "message", message, "chat", chat, "status", resp.Status)
}
