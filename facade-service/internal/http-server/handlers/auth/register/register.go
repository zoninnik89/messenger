package register

import (
	"context"
	"errors"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	pb "github.com/zoninnik89/messenger/common/api"
	grpcgateway "github.com/zoninnik89/messenger/facade-service/internal/gateway"
	"github.com/zoninnik89/messenger/facade-service/internal/lib/response"
	"github.com/zoninnik89/messenger/facade-service/internal/logging"
	"net/http"
)

type Request struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	response.Response
	Token string `json:"message_id"`
}

func New(g *grpcgateway.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.chat.login.New"
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

		loginReq := &pb.LoginRequest{
			Login:    req.Login,
			Password: req.Password,
			AppId:    1,
		}

		// Add call to GRPC handler
		res, err := g.Login(context.Background(), loginReq, requestID)

		if err != nil {
			if errors.Is(err, grpcgateway.ErrInternalServerError) {
				render.JSON(w, r, response.Error("internal server error"))
				return
			}
			render.JSON(w, r, response.Error("not valid login request"))
			return
		}

		logger.Infow("successful login", "op", op, "request_id", requestID)

		render.JSON(w, r, Response{
			Response: response.OK(),
			Token:    res.GetToken(),
		})
	}
}
