package send_message

import (
	"github.com/zoninnik89/messenger/facade-service/internal/http-server/response"
	"net/http"
)

type Request struct {
	messageText string
	chatID      string
}

type Response struct {
	response.Response
	MessageID string `json:"message_id"`
}

func New(messageSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
