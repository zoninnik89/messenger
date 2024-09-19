package main

import (
	"github.com/prometheus/common/log"
	"sync"

	pb "github.com/zoninnik89/messenger/common/api"
)

type AsyncMap struct {
	store sync.Map
}

func NewAsyncMap() *AsyncMap {
	return &AsyncMap{}
}

// Add a value to the slice at the given key
func (m *AsyncMap) Add(key string, value *Client) {
	// Use LoadOrStore to get or initialize the slice
	val, _ := m.store.LoadOrStore(key, &[]*Client{})

	// Use a mutex to protect appending to the slice
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	slice := val.(*[]*Client)      // Type assertion
	*slice = append(*slice, value) // Append to the slice
}

// Get the slice of values for the given key
func (m *AsyncMap) Get(key string) []int {
	val, ok := m.store.Load(key)
	if !ok {
		return []int{}
	}

	// Return a copy of the slice to avoid race conditions
	slice := *val.(*[]int)
	return slice
}

type Client struct {
	messageChannel chan *pb.MessageResponse
}

type PubSubService struct {
	pb.UnimplementedPubSubServiceServer
	chats *AsyncMap
}

func NewPubSubService() *PubSubService {
	return &PubSubService{chats: NewAsyncMap()}
}

func (p *PubSubService) Subscribe(req *pb.SubscribeRequest, stream pb.PubSubService_SubscribeServer) error {
	channel := make(chan *pb.MessageResponse, 500)
	client := &Client{
		messageChannel: channel,
	}
	p.chats.Add(req.Chat, client)

	for {
		select {
		case msg := <-client.messageChannel:
			if err := stream.Send(msg); err != nil {
				log.Errorf("Error sending message to client: %v", err)
				return err
			}
		case <-stream.Context().Done():
			log.Infof("Client disconnected from chat: %v", req.Chat)
			return nil
		}
	}
}

func (p *PubSub) Publish(chat chan string, msg string) {
	chat <- msg
}
