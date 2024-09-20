package main

type ChatClientInterface interface {
	SubscribeToChat(chat string)
	SendMessage(chat, message string)
}
