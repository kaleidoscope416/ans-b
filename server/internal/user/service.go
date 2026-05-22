package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ans-b/server/internal/auth"
)

var ErrUsernameTaken = errors.New("username already exists")

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	username := strings.TrimSpace(input.Username)
	nickname := strings.TrimSpace(input.Nickname)
	if username == "" {
		return nil, errors.New("username is required")
	}
	if len(username) > 64 {
		return nil, errors.New("username is too long")
	}
	if nickname != "" && len(nickname) > 100 {
		return nil, errors.New("nickname is too long")
	}
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("user service is not configured")
	}
	return s.repository.Create(ctx, CreateInput{
		Username:     username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
	})
}

func (s *Service) Profile(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user id")
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("user service is not configured")
	}
	user, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return user, err
}
