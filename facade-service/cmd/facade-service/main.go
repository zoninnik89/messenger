package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/zoninnik89/messenger/facade-service/internal/config"
	"github.com/zoninnik89/messenger/facade-service/internal/http-server/handlers/chat/send-message"
)

func main() {
	cfg := config.MustLoad()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/send", send_message.New())
}
