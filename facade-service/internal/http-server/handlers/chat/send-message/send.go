package send_message

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pb "github.com/zoninnik89/messenger/common/api"
	grpcgateway "github.com/zoninnik89/messenger/facade-service/internal/gateway"
	"github.com/zoninnik89/messenger/facade-service/internal/lib/response"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
)

type Request struct {
	MessageText string `json:"message_text" validate:"required"`
	ChatID      string `json:"chat_id" validate:"required"`
}

type Response struct {
	response.Response
	MessageID string `json:"message_id"`
	SentTS    string `json:"sent_ts"`
}

func New(g *grpcgateway.Gateway, senderID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.chat.send-message.New"
		logger := logging.GetLogger().Sugar()

		var req Request
		requestID := middleware.GetReqID(r.Context())

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logger.Errorw(
				"error while decoding request body",
				"op", op,
				"request_id", requestID,
				"err", err)

			render.JSON(w, r, response.Error("failed to decode request body"))

			return
		}

		logger.Infow("request body decoded", "op", op, "request_id", requestID, "req", req)

		if err := validator.New().Struct(req); err != nil {
			logger.Infow("invalid request", "op", op, "request_id", requestID, "request", req, "error", err)
			validateErr := err.(validator.ValidationErrors)

			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		messageID := uuid.New().String()
		sentTS := time.Now().Unix()
		senderID := senderID

		message := &pb.Message{
			MessageId:   messageID,
			ChatId:      req.ChatID,
			SenderId:    senderID,
			MessageText: req.MessageText,
			SentTs:      strconv.FormatInt(sentTS, 10),
		}

		res, err := g.SendMessage(
			context.Background(),
			&pb.SendMessageRequest{
				Message: message,
			},
		)

		if err != nil {
			if errors.Is(err, grpcgateway.ErrInternalServerError) {
				render.JSON(w, r, response.Error("internal server error"))
				return
			}
			render.JSON(w, r, response.Error("not valid message request"))
			return
		}

		logger.Infow("message sent", "op", op, "request_id", requestID, "response", res.Status)

		render.JSON(w, r, Response{
			Response:  response.OK(),
			MessageID: messageID,
			SentTS:    strconv.FormatInt(sentTS, 10),
		})
	}
}
