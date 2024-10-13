package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/zoninnik89/messenger/sso/internal/jwt"
	storagepkg "github.com/zoninnik89/messenger/sso/internal/storage"
	"github.com/zoninnik89/messenger/sso/internal/types"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Auth struct {
	logger      *zap.SugaredLogger
	usrSaver    types.UserSaver
	usrProvider types.UserProvider
	appProvider types.AppProvider
	tokenTTL    time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// NewAuthService returns a new instance of the Auth service
func NewAuthService(
	logger *zap.SugaredLogger,
	userSaver types.UserSaver,
	userProvider types.UserProvider,
	appProvider types.AppProvider,
	tokenTTL time.Duration,
) *Auth {

	return &Auth{
		logger:      logger,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if user with the given credentials exists in the system.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(
	ctx context.Context,
	login string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	a.logger.Info("attempting to login a user")

	user, err := a.usrProvider.GetUser(ctx, login)
	if err != nil {
		if errors.Is(err, storagepkg.ErrUserNotFound) {
			a.logger.Errorw("user not found", "error", err)

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		if errors.Is(err, storagepkg.ErrAppNotFound) {
			a.logger.Errorw("app not found", "error", err)

			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		a.logger.Warnw("failed to get user", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Errorw("password does not match", "error", err)
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.GetApp(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.logger.Infow("logged in", "user", user.ID)

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.logger.Warnw("failed to generate token", "error", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
//
// If user with given email already exists, returns error.
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (userID string, err error) {
	const op = "auth.RegisterNewUser"

	a.logger.Info("registering new user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.logger.Error("failed to hash password", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storagepkg.ErrUserExists) {
			a.logger.Warnw("user already exists", "error", err)

			return "", fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		a.logger.Error("failed to save user", zap.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.logger.Info("user registered")

	return id, nil
}
