package send_message

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zoninnik89/messenger/facade-service/internal/lib/response"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Request struct {
	MessageText string `json:"message_text" required:"true"`
	ChatID      string `json:"chat_id" required:"true"`
}

type Response struct {
	response.Response
	MessageID string `json:"message_id"`
	SentTS    string `json:"sent_ts"`
}

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.chat.send-message.New"
		logger := logging.GetLogger().Sugar()

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Errorw(
				"error while decoding request body",
				"op", op,
				"request_id", middleware.GetReqID(r.Context()),
				"err", err)

			render.JSON(w, r, response.Error("failed to decode request body"))

			return
		}

		log.Println("request body decoded", "op", op, "request_id", middleware.GetReqID(r.Context()), "req", req)

		if err := validator.New().Struct(req); err != nil {
			log.Println("invalid request", "op", op, "request_id", middleware.GetReqID(r.Context()), "request", req, "error", err)
			validateErr := err.(validator.ValidationErrors)

			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		messageID := uuid.New()
		sentTS := time.Now().Unix()
		// Get user ID from jwt token

		// Add call to GRPC handler

		render.JSON(w, r, Response{
			Response:  response.OK(),
			MessageID: messageID.String(),
			SentTS:    strconv.FormatInt(sentTS, 10),
		})
	}
}
