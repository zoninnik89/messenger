package tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/internal/service"
	"github.com/zoninnik89/messenger/pub-sub/internal/utils"
	mocks "github.com/zoninnik89/messenger/pub-sub/tests/mocks"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestSubscribe_MessageSent(t *testing.T) {
	// Create a mock stream using the manually created mock for grpc.ServerStream
	mockStream := mocks.NewMockServerStream()

	// Create the PubSubService and test the Subscribe method
	pubSubSvs := &service.PubSubService{
		Chats:  utils.NewAsyncMap(),
		Logger: zaptest.NewLogger(t).Sugar(),
	}

	// Prepare the subscription request
	req := &pb.SubscribeRequest{Chat: "test_chat"}

	fmt.Println("Starting Subscribe goroutine")

	go func() {
		err := pubSubSvs.Subscribe(req, mockStream)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	// Send a message after Subscribe has started
	fmt.Println("Sending message to channel")
	status := pubSubSvs.Publish(context.Background(), &pb.PublishRequest{
		Chat:    "test_chat",
		Message: "test_message",
	})
	fmt.Println("Message sent to channel")
	time.Sleep(1 * time.Second)
	// Check if the message was sent to the mock stream
	fmt.Println("Checking messages")
	assert.Equal(t, 1, len(mockStream.Messages))
	assert.Equal(t, "Message was sent", status.Status)
	assert.Equal(t, "test_message", mockStream.Messages[0].Message)

	mockStream.Close()
}

func TestSubscribe_MessageSentFailed(t *testing.T) {
	// Create a mock stream using the manually created mock for grpc.ServerStream
	mockStream := mocks.NewMockServerStream()

	// Create the PubSubService and test the Subscribe method
	pubSubSvs := &service.PubSubService{
		Chats:  utils.NewAsyncMap(),
		Logger: zaptest.NewLogger(t).Sugar(),
	}

	// Prepare the subscription request
	req := &pb.SubscribeRequest{Chat: "test_chat"}

	fmt.Println("Starting Subscribe goroutine")

	go func() {
		err := pubSubSvs.Subscribe(req, mockStream)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	// Send a message after Subscribe has started
	fmt.Println("Sending message to channel")
	status := pubSubSvs.Publish(context.Background(), &pb.PublishRequest{
		Chat:    "test_chat",
		Message: "test_message",
	})
	fmt.Println("Message sent to channel")
	time.Sleep(1 * time.Second)
	// Check if the message was sent to the mock stream
	fmt.Println("Checking messages")
	assert.Equal(t, 1, len(mockStream.Messages))
	assert.Equal(t, "Message was sent", status.Status)
	assert.Equal(t, "test_message", mockStream.Messages[0].Message)

	mockStream.Close()
}
