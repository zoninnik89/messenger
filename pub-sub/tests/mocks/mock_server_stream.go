package mocks

import (
	"context"
	pb "github.com/zoninnik89/messenger/common/api"
	"google.golang.org/grpc/metadata"
)

// MockServerStream is a mock implementation of grpc.ServerStream
type MockServerStream struct {
	Messages []*pb.MessageResponse
	Done     chan struct{}
}

func (mss *MockServerStream) Send(msg *pb.MessageResponse) error {
	mss.Messages = append(mss.Messages, msg)
	return nil
}

// SendMsg mocks the SendMsg method in grpc.ServerStream
func (mss *MockServerStream) SendMsg(m any) error {
	mss.Messages = append(mss.Messages, m.(*pb.MessageResponse))
	return nil
}

// RecvMsg mocks the RecvMsg method in grpc.ServerStream
func (mss *MockServerStream) RecvMsg(m any) error {
	return nil
}

// Context mocks the Context method in grpc.ServerStream
func (mss *MockServerStream) Context() context.Context {
	return context.Background()
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
