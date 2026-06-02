package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
)

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
	`, strings.TrimSpace(username)).Scan(&account.ID, &account.Username, &account.PasswordHash, &account.Nickname)
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
	`, strings.TrimSpace(username)).Scan(&account.ID, &account.Username, &account.PasswordHash, &account.Nickname)
	if err != nil {
		return nil, err
	}
	account.Role = RoleAdmin
	return &account, nil
}

func (r *Repository) InitAdminIfNeeded(ctx context.Context) error {
	if r == nil || r.db == nil {
		return errors.New("auth repository is not configured")
	}

	var (
		id           int64
		passwordHash string
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, password_hash
		FROM admin_users
		WHERE username = $1
	`, defaultAdminUsername).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		hash, hashErr := HashPasswordWithCost(defaultAdminPassword, 10)
		if hashErr != nil {
			return hashErr
		}
		_, execErr := r.db.ExecContext(ctx, `
			INSERT INTO admin_users (username, password_hash)
			VALUES ($1, $2)
		`, defaultAdminUsername, hash)
		return execErr
	}
	if err != nil {
		return err
	}

	if isBcryptHash(passwordHash) {
		return nil
	}

	hash, err := HashPasswordWithCost(defaultAdminPassword, 10)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE admin_users
		SET password_hash = $1,
		    updated_at = now()
		WHERE id = $2
	`, hash, id)
	return err
}

func (r *Repository) CheckAdminPasswordFormat(ctx context.Context) ([]AdminPasswordCheck, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("auth repository is not configured")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, password_hash
		FROM admin_users
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]AdminPasswordCheck, 0)
	for rows.Next() {
		var check AdminPasswordCheck
		if err := rows.Scan(&check.ID, &check.Username, &check.PasswordHash); err != nil {
			return nil, err
		}
		check.IsBcrypt = isBcryptHash(check.PasswordHash)
		results = append(results, check)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

type AdminPasswordCheck struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	IsBcrypt     bool   `json:"is_bcrypt"`
}

func isBcryptHash(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "$2a$") ||
		strings.HasPrefix(value, "$2b$") ||
		strings.HasPrefix(value, "$2y$")
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
