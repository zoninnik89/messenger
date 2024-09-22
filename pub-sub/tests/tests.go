package tests

import (
	"github.com/stretchr/testify/assert"
	pb "github.com/zoninnik89/messenger/common/api"
	"github.com/zoninnik89/messenger/pub-sub/service"
	mocks "github.com/zoninnik89/messenger/pub-sub/tests/mocks"
	"github.com/zoninnik89/messenger/pub-sub/utils"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestSubscribe_MessageSent(t *testing.T) {
	// Create a mock stream using the manually created mock for grpc.ServerStream
	mockStream := &mocks.MockServerStream{}

	// Create the PubSubService and test the Subscribe method
	pubSubSvs := &service.PubSubService{
		Chats:  utils.NewAsyncMap(),
		Logger: zaptest.NewLogger(t).Sugar(),
	}

	req := &pb.SubscribeRequest{Chat: "test_chat"}
	channel := make(chan *pb.MessageResponse, 1)
	channel <- &pb.MessageResponse{Message: "test_message"}

	client := &service.Client{MessageChannel: channel}
	pubSubSvs.Chats.Add(req.Chat, client)

	err := pubSubSvs.Subscribe(req, mockStream)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(mockStream.Messages))
	assert.Equal(t, "test_message", mockStream.Messages[0].Message)
}
