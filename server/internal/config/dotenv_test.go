package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvFilesLoadsValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(`
# comment
JWT_SECRET=dev-secret
OPENAI_MODEL="kimi-for-coding"
export QA_MIN_SCORE=0.45
`), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	unsetEnv(t, "JWT_SECRET")
	unsetEnv(t, "OPENAI_MODEL")
	unsetEnv(t, "QA_MIN_SCORE")

	if err := LoadDotEnvFiles(path); err != nil {
		t.Fatalf("load dotenv: %v", err)
	}

	if got := os.Getenv("JWT_SECRET"); got != "dev-secret" {
		t.Fatalf("expected JWT_SECRET, got %q", got)
	}
	if got := os.Getenv("OPENAI_MODEL"); got != "kimi-for-coding" {
		t.Fatalf("expected OPENAI_MODEL, got %q", got)
	}
	if got := os.Getenv("QA_MIN_SCORE"); got != "0.45" {
		t.Fatalf("expected QA_MIN_SCORE, got %q", got)
	}
}

func TestLoadDotEnvFilesDoesNotOverrideExistingEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("JWT_SECRET=file-secret\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	t.Setenv("JWT_SECRET", "shell-secret")

	if err := LoadDotEnvFiles(path); err != nil {
		t.Fatalf("load dotenv: %v", err)
	}

	if got := os.Getenv("JWT_SECRET"); got != "shell-secret" {
		t.Fatalf("expected shell env to win, got %q", got)
	}
}

func TestLoadDotEnvFilesIgnoresMissingFiles(t *testing.T) {
	if err := LoadDotEnvFiles(filepath.Join(t.TempDir(), ".env")); err != nil {
		t.Fatalf("missing dotenv should be ignored, got %v", err)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	previous, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}
