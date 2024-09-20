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

func (p *PubSubService) Publish(ctx context.Context, req *pb.PublishRequest) *pb.PublishResponse {
	clients := p.chats.Get(req.Chat)

	// If there are no available recipients
	if clients.Size() == 0 {
		return &pb.PublishResponse{Status: "No recipients"}
	}

	// Send the message to all clients
	for client := range clients.store {
		client.messageChannel <- &pb.MessageResponse{Message: req.Message}
	}

	return &pb.PublishResponse{Status: "Message was sent"}
}

func (p *PubSubService) removeClient(chat string, client *Client) {
	p.chats.Remove(chat, client)
}
