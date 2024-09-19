package main

import (
	"sync"

	pb "github.com/zoninnik89/messenger"
)

type Client struct {
	messageChannel chan pb.MessageResponse
}

type PubSubService struct {
	chats sync.Map
}

func NewPubSubService() *PubSub {
	return &PubSub{}
}

func (p *PubSub) Subscribe(chatName string) (any, bool) {
	return p.chats.Load(chatName)
}

func (p *PubSub) Publish(chat chan string, msg string) {
	chat <- msg
}
