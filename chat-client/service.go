package main

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/zoninnik89/messenger/chat-client/logging"
	producer "github.com/zoninnik89/messenger/chat-client/producer"
	pb "github.com/zoninnik89/messenger/common/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"strconv"
	"time"
)

type ChatClient struct {
	client pb.PubSubServiceClient
	logger *zap.SugaredLogger
	queue  *producer.MessageProducer
}

func NewChatClient(conn *grpc.ClientConn, q *producer.MessageProducer) *ChatClient {
	return &ChatClient{
		client: pb.NewPubSubServiceClient(conn),
		logger: logging.GetLogger().Sugar(),
		queue:  q,
	}
}

func (c *ChatClient) SubscribeToChat(senderID, chat string) {
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

		if msg.Message.SenderID != senderID {
			continue // add writing to another channel for sending through websockets
		}
	}
}

func (c *ChatClient) SendMessage(chatID, senderID, messageText string) {
	// publish message to the chat
	messageID := uuid.New().String()
	value := chatID + "," + senderID + "," + messageID + "," + messageText + "," + strconv.FormatInt(time.Now().Unix(), 10)

	deliveryChan := make(chan kafka.Event)
	err := c.queue.Publish(value, "messages", nil, deliveryChan)

	if err != nil {
		c.logger.Errorw("Error publishing click event in Kafka", zap.Error(err))
	}

	e := <-deliveryChan
	msg := e.(*kafka.Message)

	if msg.TopicPartition.Error != nil {
		c.logger.Errorw("Message was not published", "messageID", messageID, "error", msg.TopicPartition.Error)
	} else {
		c.logger.Infow("Message successfully published", "messageID", messageID, "topic", msg.TopicPartition, "time", msg.Timestamp.String())
	}
	close(deliveryChan)

	//c.logger.Infow("Message sent", "message", message, "chat", chat, "status", resp.Status)
}
