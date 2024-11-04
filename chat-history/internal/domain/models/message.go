package models

type Message struct {
	ID       string
	SenderID string
	ChatID   string
	Text     string
	SentAt   int64
}
