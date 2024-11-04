package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zoninnik89/messenger/chat-history/internal/domain/models"
	"github.com/zoninnik89/messenger/chat-history/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveMessage(
	ctx context.Context,
	messageID string,
	senderID string,
	chatID string,
	messageText string,
	sentTS string,
) error {

	const op = "storage.sqlite.SaveMessage"

	stmt, err := s.db.Prepare(
		"INSERT INTO messages (messageID, senderID, chatID, messageText, sentTS)",
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, messageID, senderID, chatID, messageText, sentTS)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return fmt.Errorf("%s: %w", op, storage.ErrMessageExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetChatMessages(
	ctx context.Context,
	chatID string,
	fromTS int64,
	toTS int64,
) ([]models.Message, error) {

	const op = "storage.sqlite.GetChatMessages"
	stmt, err := s.db.Prepare(
		"SELECT id, senderID, chatID, text, sentAt FROM messages WHERE chatID = ? AND fromTS >= ? AND toTS <= ?",
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, chatID, fromTS, toTS)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrChatNotFound)
	}

	var messages []models.Message

	for rows.Next() {
		var msg models.Message

		// Scan each row into the Message struct
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ChatID, &msg.Text, &msg.SentAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return messages, nil
}

func (s *Storage) GetChatsByUserID(ctx context.Context, userID string) ([]models.Chat, error) {
	const op = "storage.sqlite.GetChatsByUserID"

	stmt, err := s.db.Prepare(
		"SELECT id FROM chats WHERE userID = ?",
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}

	var chats []models.Chat

	for rows.Next() {
		var chat models.Chat

		// Scan each row into the Message struct
		err := rows.Scan(&chat.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return chats, nil
}
