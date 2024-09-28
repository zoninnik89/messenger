package tests

import (
	"context"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/zoninnik89/messenger/common/api"
	suite "github.com/zoninnik89/messenger/pub-sub/tests/suite"
	"testing"
	"time"
)

func TestMessageProduceConsume_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	ctx2, st2 := suite.New(t)

	messageId := gofakeit.UUID()
	chatID := gofakeit.UUID()
	senderID := gofakeit.UUID()
	messageText := gofakeit.Word()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	ctxWithCancel2, cancel2 := context.WithCancel(ctx2)

	receivedMessagesChan := make(chan *pb.Message, 10)
	receivedMessagesChan2 := make(chan *pb.Message, 10)

	go st.SubscribeToChat(ctxWithCancel, chatID, receivedMessagesChan)

	go st2.SubscribeToChat(ctxWithCancel2, chatID, receivedMessagesChan2)

	time.Sleep(2 * time.Second)

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
			} else {
				cancel2()
				receivedMessagesChan2 = nil
			}
		default:
			break
		}

		// Break the loop when both channels are closed
		if len(receivedMessages) == 1 && len(receivedMessages2) == 1 {
			cancel()
			close(receivedMessagesChan)
			cancel2()
			close(receivedMessagesChan2)
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

	messageId := gofakeit.UUID()
	chatID := gofakeit.UUID()
	senderID := gofakeit.UUID()
	messageText := gofakeit.Word()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	ctxWithCancel2, cancel2 := context.WithCancel(ctx2)

	receivedMessagesChan := make(chan *pb.Message, 10)
	receivedMessagesChan2 := make(chan *pb.Message, 10)

	go st.SubscribeToChat(ctxWithCancel, chatID, receivedMessagesChan)

	go st2.SubscribeToChat(ctxWithCancel2, chatID, receivedMessagesChan2)

	time.Sleep(2 * time.Second)

	cancel2()

	err := st.SendMessage(ctx, messageId, chatID, senderID, messageText)
	require.NoError(t, err)

	var receivedMessages []*pb.Message
	var receivedMessages2 []*pb.Message

	for i := 0; i < 10; i++ {
		select {
		case msg, ok := <-receivedMessagesChan:
			if ok {
				receivedMessages = append(receivedMessages, msg)
			} else {
				cancel()
				receivedMessagesChan = nil
			}
		case msg, ok := <-receivedMessagesChan2:
			if ok {
				receivedMessages2 = append(receivedMessages2, msg)
			} else {
				cancel2()
				receivedMessagesChan2 = nil
			}
		}
	}

	cancel()
	close(receivedMessagesChan)
	close(receivedMessagesChan2)

	assert.Equal(t, len(receivedMessages), 0)
	assert.Equal(t, len(receivedMessagesChan2), 0)
}

func TestMessageProduceConsume_NoChat(t *testing.T) {

}
