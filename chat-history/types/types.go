package types

type ChatHistoryServiceInterface interface {
	SaveMessage(chatID, message string) error
	GetMessages(chatID string) ([]string, error)
}
