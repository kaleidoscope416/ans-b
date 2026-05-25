package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

var (
	ErrInvalidLoginInput  = errors.New("username and password are required")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	Token     string  `json:"token"`
	ExpiresIn int64   `json:"expires_in"`
	User      Account `json:"user"`
}

type Service struct {
	repository *Repository
	tokens     *TokenManager
}

func NewService(repository *Repository, tokens *TokenManager) *Service {
	return &Service{repository: repository, tokens: tokens}
}

func (s *Service) LoginStudent(ctx context.Context, input LoginInput) (*LoginResult, error) {
	return s.login(ctx, input, RoleStudent)
}

func (s *Service) LoginAdmin(ctx context.Context, input LoginInput) (*LoginResult, error) {
	return s.login(ctx, input, RoleAdmin)
}

func (s *Service) login(ctx context.Context, input LoginInput, role string) (*LoginResult, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" || strings.TrimSpace(input.Password) == "" {
		return nil, ErrInvalidLoginInput
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("auth service is not configured")
	}
	if s.tokens == nil {
		return nil, errors.New("token manager is not configured")
	}

	var account *Account
	var err error
	if role == RoleAdmin {
		account, err = s.repository.FindAdminByUsername(ctx, username)
	} else {
		account, err = s.repository.FindStudentByUsername(ctx, username)
	}
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, expiresIn, err := s.tokens.Sign(account.ID, account.Username, account.Role)
	if err != nil {
		return nil, err
	}
	account.PasswordHash = ""
	return &LoginResult{Token: token, ExpiresIn: expiresIn, User: *account}, nil
}

func HashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return "", errors.New("password must be at least 6 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
