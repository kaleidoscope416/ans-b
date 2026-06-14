package auth

import (
	"context"
	"errors"
	"strings"
	"time"

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
	sessions   SessionStore
}

func NewService(repository *Repository, tokens *TokenManager, sessions SessionStore) *Service {
	return &Service{repository: repository, tokens: tokens, sessions: sessions}
}

func (s *Service) InitAuthSystem(ctx context.Context) error {
	if s == nil || s.repository == nil {
		return errors.New("auth service is not configured")
	}
	if s.tokens == nil {
		return errors.New("token manager is not configured")
	}
	if err := s.tokens.ready(); err != nil {
		return err
	}
	return s.repository.InitAdminIfNeeded(ctx)
}

func (s *Service) LoginStudent(ctx context.Context, input LoginInput) (*LoginResult, error) {
	return s.login(ctx, input, RoleStudent)
}

func (s *Service) LoginAdmin(ctx context.Context, input LoginInput) (*LoginResult, error) {
	return s.login(ctx, input, RoleAdmin)
}

func (s *Service) login(ctx context.Context, input LoginInput, role string) (*LoginResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return nil, ErrInvalidLoginInput
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("auth service is not configured")
	}
	if s.tokens == nil {
		return nil, errors.New("token manager is not configured")
	}
	if s.sessions == nil {
		return nil, errors.New("session store is not configured")
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
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	sessionID, err := NewSessionID()
	if err != nil {
		return nil, err
	}
	token, expiresIn, err := s.tokens.Sign(account.ID, account.Username, account.Role, sessionID)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	if err := s.sessions.Create(ctx, Session{
		ID:        sessionID,
		UserID:    account.ID,
		Username:  account.Username,
		Role:      account.Role,
		ExpiresAt: expiresAt,
	}, time.Duration(expiresIn)*time.Second); err != nil {
		return nil, err
	}
	account.PasswordHash = ""
	return &LoginResult{Token: token, ExpiresIn: expiresIn, User: *account}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if s == nil {
		return errors.New("auth service is not configured")
	}
	if s.tokens == nil {
		return errors.New("token manager is not configured")
	}
	if s.sessions == nil {
		return errors.New("session store is not configured")
	}
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return err
	}
	return s.sessions.Delete(ctx, claims.SessionID)
}

func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, bcrypt.DefaultCost)
}

func HashPasswordWithCost(password string, cost int) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return "", errors.New("password must be at least 6 characters")
	}
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
