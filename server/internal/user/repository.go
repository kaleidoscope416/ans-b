package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type CreateInput struct {
	Username     string
	PasswordHash string
	Nickname     string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (*User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("user repository is not configured")
	}

	var created User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (username, password_hash, nickname)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id, username, COALESCE(nickname, '')
	`, input.Username, input.PasswordHash, input.Nickname).Scan(&created.ID, &created.Username, &created.Nickname)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return &created, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("user repository is not configured")
	}

	var found User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, COALESCE(nickname, '')
		FROM users
		WHERE id = $1
	`, id).Scan(&found.ID, &found.Username, &found.Nickname)
	if err != nil {
		return nil, err
	}
	return &found, nil
}

func isUniqueViolation(err error) bool {
	message := err.Error()
	return strings.Contains(message, "duplicate key value") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "idx_users_username")
}
