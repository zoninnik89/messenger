package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zoninnik89/messenger/sso/internal/domain/models"
	"github.com/zoninnik89/messenger/sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, login string, passHash []byte) (string, error) {
	const op = "storage.sqlite.CreateUser"

	stmt, err := s.db.Prepare("INSERT INTO users (login, pass_hash, user_id) values (?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	userID := uuid.New().String()

	_, err = stmt.ExecContext(ctx, login, passHash, userID)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Storage) GetUser(ctx context.Context, login string) (models.User, error) {
	const op = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare("SELECT user_id, login, pass_hash FROM users WHERE login = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, login)

	var user models.User
	err = row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) GetApp(ctx context.Context, id int) (models.App, error) {
	const op = "storage.sqlite.GetApp"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, id)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
