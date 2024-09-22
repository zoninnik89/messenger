package main

import pb "github.com/zoninnik89/messenger/common/api"

type ChatClientInterface interface {
	SubscribeToChat(chat string)
	SendMessage(chat string, message pb.Message)
}
