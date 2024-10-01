package grpcgateway

import (
	"context"
	"errors"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/common/discovery"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

type Gateway struct {
	registry discovery.Registry
	logger   *zap.SugaredLogger
}

func NewGRPCGateway(r discovery.Registry) *Gateway {
	logger := logging.GetLogger().Sugar()
	return &Gateway{
		registry: r,
		logger:   logger,
	}
}

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrOneOfFieldsMissing  = errors.New("one of the fields are missing")
)

func (g *Gateway) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	const op = "grpcgateway.SendMessage"
	g.logger.Infow("starting connection with chat-client service")
	conn, err := discovery.ServiceConnection(ctx, "chat-client", g.registry)
	if err != nil {
		g.logger.Errorw("error while connecting to chat-client", "op", op, "error", err)
		return nil, err
	}

	g.logger.Infow("connected to chat-client")

	client := pb.NewChatClientServiceClient(conn)
	res, err := client.SendMessage(context.Background(), req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.InvalidArgument {
				g.logger.Errorw("error while sending message", "op", op, "error", err)
				return nil, ErrOneOfFieldsMissing
			}
		}
		return nil, ErrInternalServerError
	}

	return res, err
}

func (g *Gateway) GetMessagesStream(ctx context.Context, req *pb.GetMessagesStreamRequest, messages chan<- *pb.Message) error {
	const op = "grpcgateway.GetMessagesStream"
	g.logger.Infow("starting connection with chat-client service", "op", op)

	conn, err := discovery.ServiceConnection(ctx, "chat-client", g.registry)
	if err != nil {
		g.logger.Errorw("error while connecting to chat-client", "op", op, "error", err)
		err
	}

	g.logger.Infow("connected to chat-client")

	client := pb.NewChatClientServiceClient(conn)

	stream, err := client.GetMessagesStream(ctx, &pb.GetMessagesStreamRequest{UserId: userID})
	if err != nil {
		return
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
			return
		case msg, ok := <-recvChan:
			if !ok {
				// Stream has been closed
				return
			}
			// Send the received message to the messages channel
			messages <- msg
		case err := <-errChan:
			// Handle any error from stream.Recv()
			_ = err // You can log or handle the error here
			return
		}
	}

	res, err := client.SendMessage(context.Background(), req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.InvalidArgument {
				g.logger.Errorw("error while sending message", "op", op, "error", err)
				return nil, ErrOneOfFieldsMissing
			}
		}
		return nil, ErrInternalServerError
	}

	return res, err
}

func (g *Gateway) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	const op = "grpcgateway.Login"
	g.logger.Infow("starting connection with sso service")

}

func (g *Gateway) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	const op = "grpcgateway.Register"
	g.logger.Infow("starting connection with sso service")

}
