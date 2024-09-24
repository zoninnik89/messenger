package main

import (
	"github.com/zoninnik89/messenger/sso/internal/config"
	"github.com/zoninnik89/messenger/sso/internal/logging"
)

func main() {
	cfg := config.MustLoad()
	logger := logging.InitLogger()

	logger.Info("Starting sso service")
}
