package types

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "github.com/zoninnik89/messenger/common/api"
)

type PubSubServiceInterface interface {
	Subscribe(chatID string, stream pb.PubSubService_SubscribeServer) error
	ConsumeMessage(ctx context.Context, consumer *kafka.Consumer) (string, error)
}

type Client struct {
	MessageChannel *chan *pb.MessageResponse
}
