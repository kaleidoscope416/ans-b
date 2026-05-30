package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMiddlewareAllowsStoredSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewTokenManager("test-secret", time.Hour)
	store := NewMemorySessionStore()
	token := createTestSessionToken(t, manager, store, "session-1")

	engine := gin.New()
	engine.GET("/protected", Middleware(manager, store, RoleStudent), func(c *gin.Context) {
		claims, ok := CurrentUser(c)
		if !ok {
			t.Fatal("expected claims in context")
		}
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected stored session to pass, got %d", recorder.Code)
	}
}

func TestMiddlewareRejectsMissingSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewTokenManager("test-secret", time.Hour)
	token, _, err := manager.Sign(12, "student1", RoleStudent, "missing-session")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	engine := gin.New()
	engine.GET("/protected", Middleware(manager, NewMemorySessionStore(), RoleStudent), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing session to fail with 401, got %d", recorder.Code)
	}
}

func TestMiddlewareRejectsSessionStoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewTokenManager("test-secret", time.Hour)
	token, _, err := manager.Sign(12, "student1", RoleStudent, "session-1")
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	engine := gin.New()
	engine.GET("/protected", Middleware(manager, failingSessionStore{}, RoleStudent), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected store error to fail with 503, got %d", recorder.Code)
	}
}

func TestLogoutDeletesCurrentSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := NewTokenManager("test-secret", time.Hour)
	store := NewMemorySessionStore()
	token := createTestSessionToken(t, manager, store, "session-1")
	handler := NewHandler(NewService(nil, manager, store))

	engine := gin.New()
	handler.RegisterRoutes(engine.Group("/auth"))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected logout to return 200, got %d", recorder.Code)
	}
	if _, err := store.Get(context.Background(), "session-1"); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected session to be deleted, got %v", err)
	}
}

func createTestSessionToken(t *testing.T, manager *TokenManager, store SessionStore, sessionID string) string {
	t.Helper()

	token, _, err := manager.Sign(12, "student1", RoleStudent, sessionID)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if err := store.Create(context.Background(), Session{
		ID:        sessionID,
		UserID:    12,
		Username:  "student1",
		Role:      RoleStudent,
		ExpiresAt: time.Now().Add(time.Hour),
	}, time.Hour); err != nil {
		t.Fatalf("create session: %v", err)
	}
	return token
}

type failingSessionStore struct{}

func (failingSessionStore) Create(context.Context, Session, time.Duration) error {
	return errors.New("store unavailable")
}

func (failingSessionStore) Get(context.Context, string) (*Session, error) {
	return nil, errors.New("store unavailable")
}

func (failingSessionStore) Delete(context.Context, string) error {
	return errors.New("store unavailable")
}
