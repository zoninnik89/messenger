package main

import (
	"context"
	"fmt"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc"
	"io"
	"log"
)

type ChatClient struct {
	client pb.PubSubServiceClient
}

func NewChatClient(conn *grpc.ClientConn) *ChatClient {
	return &ChatClient{
		client: pb.NewPubSubServiceClient(conn),
	}
}

func (c *ChatClient) SubscribeToChat(chat string) {
	stream, err := c.client.Subscribe(context.Background(), &pb.SubscribeRequest{Chat: chat})
	if err != nil {
		log.Fatalf("Failed to subscribe to chat: %v", err)
	}

	log.Printf("Subscribed to chat: %v", chat)

	// Continuously receive messages from the stream
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			// Server closed connection
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a message: %v", err)
		}

		// Log the received message
		fmt.Printf("Received a message: %v from: %v", msg.Message, chat)
	}
}

func (c *ChatClient) SendMessage(chat, message string) {
	// publish message to the chat
	resp, err := c.client.Publish(context.Background(), &pb.PublishRequest{
		Chat:    chat,
		Message: message,
	})
	if err != nil {
		log.Fatalf("Failed to send a message: %v", err)
	}

	log.Printf("Meesage was sent: %v, Status: %v", message, resp.Status)
}
