package storage

import (
	"errors"
)

var (
	ErrChatExists    = errors.New("chat already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrChatNotFound  = errors.New("chat not found")
	ErrMessageExists = errors.New("message already exists")
)
