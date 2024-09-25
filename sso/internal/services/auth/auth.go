package auth

import "go.uber.org/zap"

type Auth struct {
	logger *zap.SugaredLogger
}

func NewAuthService(logger *zap.SugaredLogger) *Auth {}
