package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

// Global logger instance
var (
	logger *zap.Logger
	once   sync.Once
)

func InitLogger() *zap.Logger {
	// Initialize a logger exactly once
	once.Do(func() {
		// Customize Zap logger
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		var err error
		logger, err = config.Build()
		if err != nil {
			panic(err)
		}
	})
	return logger
}

func GetLogger() *zap.Logger {
	if logger == nil {
		return InitLogger()
	}
	return logger
}

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
