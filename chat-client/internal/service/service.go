package service

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zoninnik89/messenger/chat-client/internal/logging"
	"github.com/zoninnik89/messenger/chat-client/internal/producer"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/common/discovery"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io"
)

type ChatClient struct {
	logger   *zap.SugaredLogger
	queue    *producer.MessageProducer
	registry discovery.Registry
}

func NewChatClient(r discovery.Registry, q *producer.MessageProducer) (*ChatClient, error) {
	const op = "service.NewChatClient"
	logger := logging.GetLogger().Sugar()

	return &ChatClient{
		logger:   logger,
		queue:    q,
		registry: r,
	}, nil
}

func (c *ChatClient) SubscribeForMessages(
	ctx context.Context,
	userID string,
	stream pb.ChatClientService_GetMessagesStreamServer,
) error {

	const op = "service.SubscribeToChat"

	c.logger.Infow("starting connection with pub-sub service", "op", op)

	conn, err := discovery.ServiceConnection(context.Background(), "pub-sub", c.registry)

	if err != nil {
		c.logger.Fatalw("failed to dial pub-sub", "op", op, "err", err)
		return err
	}

	client := pb.NewPubSubServiceClient(conn)

	c.logger.Infow("connected to pub-sub service")

	subscribeRequest := &pb.SubscribeRequest{
		UserId: userID,
	}

	c.logger.Infow("subscribing to messages", "op", op, "user ID", userID)

	streamFromPubSub, err := client.Subscribe(ctx, subscribeRequest)
	if err != nil {
		c.logger.Fatalw("failed to subscribe to pub-sub", "op", op, "user ID", userID, "err", err)
	}

	c.logger.Infow("subscribed to pub-sub", "op", op, "user", userID)

	// Create a channel to receive stream messages or errors
	recvChan := make(chan *pb.Message)
	errChan := make(chan error)

	// Start a goroutine to receive messages from the stream
	go func() {
		defer func() {
			c.logger.Infow("closing connection", "op", op, "user", userID)
			close(recvChan)
			close(errChan)
		}()
		for {
			msg, err := streamFromPubSub.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				errChan <- err
				return
			}
			recvChan <- msg
		}
	}()

	for {
		select {
		case <-stream.Context().Done():
			c.logger.Infow("user disconnected from server", "op", op, "user ID", userID)
			return nil
		case <-ctx.Done():
			// Context canceled, stop receiving messages
			c.logger.Infow("user disconnected from server", "op", op, "user ID", userID)
			return nil
		case msg, ok := <-recvChan:
			if !ok {
				// Stream has been closed
				return nil
			}
			if err := stream.Send(msg); err != nil {
				c.logger.Errorw("error sending message to user", "op", op, "err", err)
				return fmt.Errorf("%s: %w", op, err)
			}
			c.logger.Infow("sent message to user", "op", op, "user", userID)
		case err := <-errChan:
			// Handle any error from stream.Recv()
			c.logger.Errorw("error receiving message from pub-sub", "op", op, "err", err)
			return nil
		}
	}
}

func (c *ChatClient) SendMessage(
	messageID string,
	chatID string,
	senderID string,
	messageText string,
	sentTime string,
) error {
	const op = "service.SendMessage"

	// publish message to the chat
	message := &pb.Message{
		MessageId:   messageID,
		SenderId:    senderID,
		ChatId:      chatID,
		MessageText: messageText,
		SentTs:      sentTime,
	}

	serializedMessage, err := proto.Marshal(message)
	if err != nil {
		c.logger.Errorw("failed to serialize message", "op", op, "messageID", messageID, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err = c.queue.Publish(serializedMessage, "messages", nil, deliveryChan)

	if err != nil {
		c.logger.Errorw("failed to publish message in Kafka", "op", op, "messageID", messageID, "err", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	e := <-deliveryChan
	msg := e.(*kafka.Message)

	if msg.TopicPartition.Error != nil {
		c.logger.Errorw("message was not published", "messageID", messageID, "error", msg.TopicPartition.Error)
		return fmt.Errorf("%s: %w", op, msg.TopicPartition.Error)
	} else {
		c.logger.Infow("message successfully published", "messageID", messageID, "topic", msg.TopicPartition, "time", msg.Timestamp.String())
	}

	return nil
}
