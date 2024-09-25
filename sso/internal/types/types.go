package types

import (
	"context"
	"github.com/zoninnik89/messenger/sso/internal/domain/models"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	Register(ctx context.Context, email string, password string) (userID int64, err error)
}

type Storage interface {
	SaveUser(ctx context.Context) (uid int64, err error)
	UpdateUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, email string) (models.User, error)
}
