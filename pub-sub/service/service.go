package main

import (
	"context"

	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/logging"
	"go.uber.org/zap"
)

type Client struct {
	messageChannel chan *pb.MessageResponse
}

type PubSubService struct {
	chats  *AsyncMap
	logger *zap.SugaredLogger
}

func NewPubSubService() *PubSubService {
	return &PubSubService{chats: NewAsyncMap(), logger: logging.GetLogger().Sugar()}
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
				p.logger.Errorw("error sending message to client", "err", err)
				return err
			}
		case <-stream.Context().Done():
			p.logger.Infow("Client disconnected from chat", "chat", req.Chat)
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
