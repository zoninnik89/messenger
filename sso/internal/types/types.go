package types

import (
	"context"
	"github.com/zoninnik89/messenger/sso/internal/domain/models"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	//UpdateUser(ctx context.Context, user models.User) error
}

type UserProvider interface {
	GetUser(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}
