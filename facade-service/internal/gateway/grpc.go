package grpcgateway

import (
	"context"
	"errors"
	"fmt"
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
	ErrUserIDIsMissing     = errors.New("user ID is missing")
)

// SendMessage method establishes GRPC connection with Chat-client service and makes a request to send a message.
func (g *Gateway) SendMessage(ctx context.Context, req *pb.SendMessageRequest, requestID string) (*pb.SendMessageResponse, error) {
	const op = "grpcgateway.SendMessage"
	g.logger.Infow("starting connection with chat-client service", "request ID", requestID)
	conn, err := discovery.ServiceConnection(ctx, "chat-client", g.registry)
	if err != nil {
		g.logger.Errorw("error while connecting to chat-client", "op", op, "requestID", requestID, "req", req, "error", err)
		return nil, err
	}

	g.logger.Infow("connected to chat-client")

	client := pb.NewChatClientServiceClient(conn)
	res, err := client.SendMessage(context.Background(), req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.InvalidArgument {
				g.logger.Errorw("error while sending message", "op", op, "requestID", requestID, "req", req, "error", err)
				return nil, fmt.Errorf("%s: %s", op, st.Message())
			}
		}
		return nil, ErrInternalServerError
	}

	return res, err
}

// GetMessagesStream method establishes persistent GRPC connection with Chat-client service and gets a stream of messages
// for all chats, where the given user is on participants.
func (g *Gateway) GetMessagesStream(ctx context.Context, req *pb.GetMessagesStreamRequest, messages chan<- *pb.Message, requestID string) error {
	const op = "grpcgateway.GetMessagesStream"
	g.logger.Infow("starting connection with chat-client service", "op", op, "requestID", requestID)

	conn, err := discovery.ServiceConnection(ctx, "chat-client", g.registry)
	if err != nil {
		g.logger.Errorw("error while connecting to chat-client", "op", op, "requestID", requestID, "req", req, "error", err)
		return err
	}

	g.logger.Infow("connected to chat-client")

	client := pb.NewChatClientServiceClient(conn)

	stream, err := client.GetMessagesStream(ctx, &pb.GetMessagesStreamRequest{UserId: req.GetUserId()})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.InvalidArgument {
				return ErrUserIDIsMissing
			}
		}
		return ErrInternalServerError
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
			g.logger.Errorw("error while receiving message", "op", op, "requestID", requestID, "userID", req.GetUserId(), "error", err)
			return err
		}
	}

}

// Login method establishes GRPC connection with SSO service and makes a request to log in a user.
func (g *Gateway) Login(ctx context.Context, req *pb.LoginRequest, requestID string) (*pb.LoginResponse, error) {
	const op = "grpcgateway.Login"
	g.logger.Infow("starting connection with sso service")

	conn, err := discovery.ServiceConnection(ctx, "sso", g.registry)
	if err != nil {
		g.logger.Errorw("error while connecting to sso service", "op", op, "requestID", requestID, "req", req, "error", err)
		return nil, err
	}

	g.logger.Infow("connected to sso service")

	client := pb.NewAuthServiceClient(conn)
	res, err := client.Login(context.Background(), req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.InvalidArgument {
				g.logger.Errorw("error while logging in", "op", op, "requestID", requestID, "req", req, "error", err)
				return nil, fmt.Errorf("%s: %s", op, st.Message())
			}
		}
		return nil, ErrInternalServerError
	}

	return res, err
}

// Register method establishes GRPC connection with SSO service and makes a request to register a user.
func (g *Gateway) Register(ctx context.Context, req *pb.RegisterRequest, requestID string) (*pb.RegisterResponse, error) {
	const op = "grpcgateway.Register"
	g.logger.Infow("starting connection with sso service", "op", op, "requestID", requestID)

	conn, err := discovery.ServiceConnection(ctx, "sso-service", g.registry)
	if err != nil {
		g.logger.Errorw("error while connecting to sso service", "op", op, "requestID", requestID, "error", err)
		return nil, err
	}

	g.logger.Infow("connected to sso service", "op", op, "requestID", requestID)

	client := pb.NewAuthServiceClient(conn)
	res, err := client.Register(context.Background(), req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.InvalidArgument {
				g.logger.Errorw("error while registering", "op", op, "requestID", requestID, "req", req, "error", err)
				return nil, fmt.Errorf("%s: %s", op, st.Message())
			}
		}
		return nil, ErrInternalServerError
	}

	return res, err
}
