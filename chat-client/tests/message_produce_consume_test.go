package tests

import (
	"context"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	suite "github.com/zoninnik89/messenger/chat-client/tests/suite"
	pb "github.com/zoninnik89/messenger/common/api"
	"testing"
	"time"
)

func TestMessageSendReceive_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	ctx2, st2 := suite.New(t)

	userID := "user1"
	userID2 := "user2"

	messageId := gofakeit.UUID()
	chatID := "chat1"
	senderID := "user3"
	messageText := gofakeit.Word()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	ctxWithCancel2, cancel2 := context.WithCancel(ctx2)

	receivedMessagesChan := make(chan *pb.Message, 10)
	receivedMessagesChan2 := make(chan *pb.Message, 10)

	go st.SubscribeToChat(ctxWithCancel, userID, receivedMessagesChan)

	go st2.SubscribeToChat(ctxWithCancel2, userID2, receivedMessagesChan2)

	time.Sleep(3 * time.Second)

	err := st.SendMessage(ctx, messageId, chatID, senderID, messageText)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	var receivedMessages []*pb.Message
	var receivedMessages2 []*pb.Message

	for {
		select {
		case msg, ok := <-receivedMessagesChan:
			if ok {
				receivedMessages = append(receivedMessages, msg)
			}
		case msg, ok := <-receivedMessagesChan2:
			if ok {
				receivedMessages2 = append(receivedMessages2, msg)
			}
		case <-time.After(5 * time.Second):
			cancel()
			cancel2()
			break
		}

		// Break the loop when both channels are closed
		if len(receivedMessages) == 1 && len(receivedMessages2) == 1 {
			cancel()
			cancel2()
			break
		}
	}

	assert.Equal(t, len(receivedMessages), 1)
	assert.Equal(t, receivedMessages[0].MessageId, messageId)
	assert.Equal(t, receivedMessages[0].ChatId, chatID)
	assert.Equal(t, receivedMessages[0].SenderId, senderID)
	assert.Equal(t, receivedMessages[0].MessageText, messageText)
}

func TestMessageProduceConsume_OneChatParticipant(t *testing.T) {
	ctx, st := suite.New(t)
	ctx2, st2 := suite.New(t)

	userID := "user1"
	userID2 := "user2"

	messageId := gofakeit.UUID()
	chatID := "chat1"
	senderID := gofakeit.UUID()
	messageText := gofakeit.Word()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	ctxWithCancel2, cancel2 := context.WithCancel(ctx2)

	receivedMessagesChan := make(chan *pb.Message, 10)
	receivedMessagesChan2 := make(chan *pb.Message, 10)

	go st.SubscribeToChat(ctxWithCancel, userID, receivedMessagesChan)

	go st2.SubscribeToChat(ctxWithCancel2, userID2, receivedMessagesChan2)

	time.Sleep(1 * time.Second)

	cancel2()

	time.Sleep(1 * time.Second)

	err := st.SendMessage(ctx, messageId, chatID, senderID, messageText)
	require.NoError(t, err)

	var receivedMessages []*pb.Message
	var receivedMessages2 []*pb.Message

	for {
		select {
		case msg, ok := <-receivedMessagesChan:
			if ok {
				receivedMessages = append(receivedMessages, msg)
			}
		case msg, ok := <-receivedMessagesChan2:
			if ok {
				receivedMessages2 = append(receivedMessages2, msg)
			}
		case <-time.After(5 * time.Second):
			cancel()
			receivedMessagesChan = nil
			receivedMessagesChan2 = nil
			break
		}

		if receivedMessagesChan == nil && receivedMessagesChan2 == nil {
			cancel()
			break
		}
	}

	assert.Equal(t, len(receivedMessages), 1)
	assert.Equal(t, len(receivedMessagesChan2), 0)
}
