package main

import (
	"context"
	"github.com/prometheus/common/log"
	pb "github.com/zoninnik89/messenger/common/api"
)

type Client struct {
	messageChannel chan *pb.MessageResponse
}

type PubSubService struct {
	pb.UnimplementedPubSubServiceServer
	chats *AsyncMap
}

func NewPubSubService() *PubSubService {
	return &PubSubService{chats: NewAsyncMap()}
}

func (p *PubSubService) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	channel := make(chan *pb.MessageResponse, 500)
	client := &Client{
		messageChannel: channel,
	}
	p.chats.Add(req.Chat, client)

	for {
		select {
		case msg := <-client.messageChannel:
			if err := stream.Send(msg); err != nil {
				log.Errorf("Error sending message to client: %v", err)
				return err
			}
		case <-stream.Context().Done():
			log.Infof("Client disconnected from chat: %v", req.Chat)
			return nil
		}
	}
}

func (p *PubSubService) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	clients := p.chats.Get(req.Chat)

	// If there are no available recipients
	if len(clients) == 0 {
		return &pb.PublishResponse{Status: "No recipients"}, nil
	}

	// Send the message to all clients
	for _, client := range clients {
		client.messageChannel <- &pb.MessageResponse{Message: req.Message}
	}

	return &pb.PublishResponse{Status: "Message was sent"}, nil
}

func (p *PubSubService) removeClient(chat string, client *Client) {
	clients := p.chats.Get(chat)

	for i, c := range clients {
		if c == client {
			p.chats
		}
	}
}
