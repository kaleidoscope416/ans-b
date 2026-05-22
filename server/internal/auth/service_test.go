package auth

import "testing"

func TestHashPasswordUsesBcrypt(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "secret123" {
		t.Fatal("password was stored in plain text")
	}
	if len(hash) < 20 {
		t.Fatalf("hash is unexpectedly short: %q", hash)
	}
}

func TestHashPasswordRejectsShortPassword(t *testing.T) {
	if _, err := HashPassword("123"); err == nil {
		t.Fatal("expected short password to fail")
	}
}
