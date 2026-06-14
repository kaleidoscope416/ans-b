package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "auth:session:"

var ErrSessionNotFound = errors.New("login session not found")

type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionStore interface {
	Create(ctx context.Context, session Session, ttl time.Duration) error
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
}

func NewSessionID() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStoreFromEnv() (*RedisSessionStore, error) {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		addr = "localhost:6379"
	}

	db := 0
	if value := strings.TrimSpace(os.Getenv("REDIS_DB")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return nil, errors.New("REDIS_DB must be a non-negative integer")
		}
		db = parsed
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	return &RedisSessionStore{client: client}, nil
}

func (s *RedisSessionStore) Ping(ctx context.Context) error {
	if s == nil || s.client == nil {
		return errors.New("redis session store is not configured")
	}
	return s.client.Ping(ctx).Err()
}

func (s *RedisSessionStore) Close() error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *RedisSessionStore) Create(ctx context.Context, session Session, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return errors.New("redis session store is not configured")
	}
	if session.ID == "" {
		return errors.New("session id is required")
	}
	if ttl <= 0 {
		return errors.New("session ttl must be positive")
	}
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, sessionKey(session.ID), data, ttl).Err()
}

func (s *RedisSessionStore) Get(ctx context.Context, id string) (*Session, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("redis session store is not configured")
	}
	data, err := s.client.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, id string) error {
	if s == nil || s.client == nil {
		return errors.New("redis session store is not configured")
	}
	if strings.TrimSpace(id) == "" {
		return nil
	}
	return s.client.Del(ctx, sessionKey(id)).Err()
}

func sessionKey(id string) string {
	return sessionKeyPrefix + id
}

type MemorySessionStore struct {
	mu       sync.Mutex
	sessions map[string]Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]Session)}
}

func (s *MemorySessionStore) Create(_ context.Context, session Session, _ time.Duration) error {
	if s == nil {
		return errors.New("memory session store is not configured")
	}
	if session.ID == "" {
		return errors.New("session id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

func (s *MemorySessionStore) Get(_ context.Context, id string) (*Session, error) {
	if s == nil {
		return nil, errors.New("memory session store is not configured")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok || !time.Now().Before(session.ExpiresAt) {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

func (s *MemorySessionStore) Delete(_ context.Context, id string) error {
	if s == nil {
		return errors.New("memory session store is not configured")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}
