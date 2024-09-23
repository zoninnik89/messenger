package service

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/pub-sub/types"
	"strings"

	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/logging"
	"github.com/zoninnik89/messenger/pub-sub/utils"
	"go.uber.org/zap"
)

type PubSubService struct {
	Chats  *utils.AsyncMap
	Logger *zap.SugaredLogger
}

func NewPubSubService() *PubSubService {
	return &PubSubService{Chats: utils.NewAsyncMap(), Logger: logging.GetLogger().Sugar()}
}

func (p *PubSubService) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	p.Logger.Infow("Subscribe method called", "chat", req.ChatID)
	channel := make(chan *pb.MessageResponse, 500)
	client := &types.Client{
		MessageChannel: &channel,
	}
	p.Chats.Add(req.ChatID, client)

	for {
		select {
		case msg := <-*client.MessageChannel:
			p.Logger.Infow("Received message from client channel", "message", msg.Message)
			if err := stream.Send(msg); err != nil {
				p.Logger.Errorw("error sending message to client", "err", err)
				p.removeClient(req.ChatID, client) // Removing the client from the Async map
				return err
			}
		case <-stream.Context().Done():
			p.Logger.Infow("Client disconnected from chat", "chat", req.ChatID)
			p.removeClient(req.ChatID, client) // Removing the client from the Async map
			return nil
		default:
			//p.Logger.Debugw("Waiting in select block")
		}
	}
}

func (p *PubSubService) ConsumeMessage(ctx context.Context, consumer *kafka.Consumer) (*pb.PublishResponse, error) {
	msg, err := consumer.ReadMessage(-1)
	if err != nil {
		p.Logger.Fatalw("Failed to read message", "err", err)
		return nil, err
	}
	msgSlice := strings.Split(string(msg.Value), ",")
	chatID, senderID, messageID, messageText, sentTime := msgSlice[0], msgSlice[1], msgSlice[2], msgSlice[3], msgSlice[4]

	clients := p.Chats.Get(chatID)

	// If there are no available recipients
	if clients.Size() == 0 {
		return &pb.PublishResponse{Status: "No recipients"}, nil
	}

	// Send the message to all clients
	for client := range clients.Store {
		p.Logger.Infow("Sending message to client", "chatID", chatID, "senderID", senderID, "messageID", messageID, "messageText", messageText, "sentTime", sentTime)
		*client.MessageChannel <- &pb.MessageResponse{Message: &pb.Message{
			SenderID:    senderID,
			MessageID:   messageID,
			MessageText: messageText,
			SentTS:      sentTime,
		}}
	}

	return &pb.PublishResponse{Status: "Message was sent to all recipients"}, nil
}

//func (p *PubSubService) Publish(ctx context.Context, req *pb.PublishRequest) *pb.PublishResponse {
//clients := p.Chats.Get(req.Chat)

// If there are no available recipients
//if clients.Size() == 0 {
//	return &pb.PublishResponse{Status: "No recipients"}
//}

// Send the message to all clients
//for client := range clients.Store {
//	p.Logger.Infow("Sending message to client", "chat", req.Chat, "message", req.Message)
//	*client.MessageChannel <- &pb.MessageResponse{Message: req.Message}
//}

//return &pb.PublishResponse{Status: "Message was sent"}
//}

func (p *PubSubService) removeClient(chat string, client *types.Client) {
	p.Chats.Remove(chat, client)
}
