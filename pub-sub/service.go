package main

import "sync"

type PubSub struct {
	chats sync.Map
}

func NewPubSub() *PubSub {
	return &PubSub{}
}

func (p *PubSub) Subscribe(chatName string) (any, bool) {
	return p.chats.Load(chatName)
}

func (p *PubSub) Publish(chat chan string, msg string) {
	chat <- msg
}
