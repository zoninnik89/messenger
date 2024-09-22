package mocks

import (
	"context"
	"fmt"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc/metadata"
)

// MockServerStream is a mock implementation of grpc.ServerStream
type MockServerStream struct {
	Messages []*pb.MessageResponse
	DoneChan chan struct{}
}

func NewMockServerStream() *MockServerStream {
	return &MockServerStream{
		Messages: make([]*pb.MessageResponse, 0),
		DoneChan: make(chan struct{}), // We control when the context is done
	}
}

func (mss *MockServerStream) Send(msg *pb.MessageResponse) error {
	fmt.Println("MockServerStream.Send called with message:", msg.Message)
	mss.Messages = append(mss.Messages, msg)
	return nil
}

// SendMsg mocks the SendMsg method in grpc.ServerStream
func (mss *MockServerStream) SendMsg(m any) error {
	fmt.Println("Message sent:", m)
	mss.Messages = append(mss.Messages, m.(*pb.MessageResponse))
	return nil
}

// RecvMsg mocks the RecvMsg method in grpc.ServerStream
func (mss *MockServerStream) RecvMsg(m any) error {
	return nil
}

// Context mocks the Context method in grpc.ServerStream
func (mss *MockServerStream) Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-mss.DoneChan // Wait for the signal to cancel the context
		cancel()
	}()
	return ctx
}

// SetHeader is a mock for grpc.ServerStream's SetHeader method
func (mss *MockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

// SendHeader is a mock for grpc.ServerStream's SendHeader method
func (mss *MockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

// SetTrailer is a mock for grpc.ServerStream's SetTrailer method
func (mss *MockServerStream) SetTrailer(md metadata.MD) {}

// Close the stream manually
func (mss *MockServerStream) Close() {
	close(mss.DoneChan) // Signal the stream is done (close context)
}
