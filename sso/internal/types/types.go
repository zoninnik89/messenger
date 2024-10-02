package types

import (
	"context"
	"github.com/zoninnik89/messenger/sso/internal/domain/models"
)

type Auth interface {
	Login(ctx context.Context, login string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, login string, password string) (userID string, err error)
}

type UserSaver interface {
	SaveUser(ctx context.Context, login string, passHash []byte) (uid string, err error)
	//UpdateUser(ctx context.Context, user models.User) error
}

type UserProvider interface {
	GetUser(ctx context.Context, login string) (models.User, error)
}

type AppProvider interface {
	GetApp(ctx context.Context, id int) (models.App, error)
}
