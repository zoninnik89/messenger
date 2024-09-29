package grpc

import (
	"context"
	"github.com/zoninnik89/messenger/chat-client/internal/logging"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/common/discovery"
	"go.uber.org/zap"
	"io"
	"log"
)

type Gateway struct {
	registry discovery.Registry
	logger   *zap.SugaredLogger
}

func NewGRPCGateway(r discovery.Registry) *Gateway {
	return &Gateway{
		registry: r,
		logger:   logging.GetLogger().Sugar(),
	}
}

func (g *Gateway) SubscribeForMessages(ctx context.Context, userID string, messages chan<- *pb.Message) error {
	const op = "gateway.SubscribeForMessages"
	g.logger.Infow("starting connection with pub-sub service", "op", op, "user ID", userID)
	conn, err := discovery.ServiceConnection(ctx, "pub-sub", g.registry)
	if err != nil {
		g.logger.Fatalw("failed to dial server", "op", op, "err", err)
	}

	g.logger.Infow("connected to pub-sub service")
	client := pb.NewPubSubServiceClient(conn)
	subscribeRequest := &pb.SubscribeRequest{
		UserId: userID,
	}
	stream, err := client.Subscribe(ctx, subscribeRequest)
	if err != nil {
		g.logger.Fatalw("failed to subscribe to pub-sub", "op", op, "err", err)
	}

	defer func() {
		log.Println("closing connection")
		stream.CloseSend()
		close(messages)
	}()

	// Create a channel to receive stream messages or errors
	recvChan := make(chan *pb.Message)
	errChan := make(chan error)

	// Start a goroutine to receive messages from the stream
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					close(recvChan)
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
		case <-ctx.Done():
			// Context canceled, stop receiving messages
			return nil
		case msg, ok := <-recvChan:
			if !ok {
				// Stream has been closed
				return nil
			}
			// Send the received message to the messages channel
			messages <- msg
		case err := <-errChan:
			// Handle any error from stream.Recv()
			_ = err // You can log or handle the error here
			return nil
		}
	}

}
