package auth

import (
	"context"
	"database/sql"
	"errors"
)

type Account struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Nickname     string `json:"nickname"`
	Role         string `json:"role"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindStudentByUsername(ctx context.Context, username string) (*Account, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("auth repository is not configured")
	}

	var account Account
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, COALESCE(nickname, '')
		FROM users
		WHERE username = $1
	`, username).Scan(&account.ID, &account.Username, &account.PasswordHash, &account.Nickname)
	if err != nil {
		return nil, err
	}
	account.Role = RoleStudent
	return &account, nil
}

func (r *Repository) FindAdminByUsername(ctx context.Context, username string) (*Account, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("auth repository is not configured")
	}

	var account Account
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, username
		FROM admin_users
		WHERE username = $1
	`, username).Scan(&account.ID, &account.Username, &account.PasswordHash, &account.Nickname)
	if err != nil {
		return nil, err
	}
	account.Role = RoleAdmin
	return &account, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
