package auth

import (
	"testing"
	"time"
)

func TestTokenManagerSignsAndVerifiesToken(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)

	token, expiresIn, err := manager.Sign(12, "student1", RoleStudent)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if expiresIn != 3600 {
		t.Fatalf("expected 3600 seconds, got %d", expiresIn)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID != 12 || claims.Username != "student1" || claims.Role != RoleStudent {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestTokenManagerRejectsTamperedToken(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	token, _, err := manager.Sign(12, "student1", RoleStudent)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := manager.Verify(token + "x"); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestBearerTokenParsesAuthorizationHeader(t *testing.T) {
	token, err := BearerToken("Bearer abc.def")
	if err != nil {
		t.Fatalf("bearer token: %v", err)
	}
	if token != "abc.def" {
		t.Fatalf("unexpected token %q", token)
	}
}
