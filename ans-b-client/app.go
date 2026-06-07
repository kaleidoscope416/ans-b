package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const defaultAPIBaseURL = "http://127.0.0.1:23456"

// App struct
type App struct {
	ctx          context.Context
	apiBaseURL   string
	httpClient   *http.Client
	mu           sync.Mutex
	studentToken string
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type Account struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

type LoginResult struct {
	Token     string  `json:"token"`
	ExpiresIn int64   `json:"expires_in"`
	User      Account `json:"user"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	baseURL := strings.TrimSpace(os.Getenv("ANS_B_API_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	return &App{
		apiBaseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LoginStudent(username string, password string) (*LoginResult, error) {
	log.Printf("wails proxy login start username=%q", strings.TrimSpace(username))
	var result LoginResult
	if err := a.postJSON("/api/v1/auth/student/login", map[string]any{
		"username": strings.TrimSpace(username),
		"password": strings.TrimSpace(password),
	}, false, &result); err != nil {
		log.Printf("wails proxy login error username=%q error=%v", strings.TrimSpace(username), err)
		return nil, err
	}
	a.mu.Lock()
	a.studentToken = result.Token
	a.mu.Unlock()
	log.Printf("wails proxy login done username=%q", strings.TrimSpace(username))
	return &result, nil
}

func (a *App) RegisterStudent(username string, password string, nickname string) (*LoginResult, error) {
	log.Printf("wails proxy register start username=%q", strings.TrimSpace(username))
	if err := a.postJSON("/api/v1/users/register", map[string]any{
		"username": strings.TrimSpace(username),
		"password": strings.TrimSpace(password),
		"nickname": strings.TrimSpace(nickname),
	}, false, nil); err != nil {
		log.Printf("wails proxy register error username=%q error=%v", strings.TrimSpace(username), err)
		return nil, err
	}
	log.Printf("wails proxy register done username=%q", strings.TrimSpace(username))
	result, err := a.LoginStudent(username, password)
	if err != nil {
		return nil, fmt.Errorf("注册成功，但自动登录失败：%w", err)
	}
	return result, nil
}

func (a *App) AskQuestion(question string, limit int) (map[string]any, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("question is required")
	}
	if limit <= 0 {
		limit = 5
	}
	startedAt := time.Now()
	log.Printf("wails proxy ask start question=%q limit=%d", question, limit)
	var result map[string]any
	if err := a.postJSON("/api/v1/qa/ask", map[string]any{
		"question": question,
		"limit":    limit,
	}, true, &result); err != nil {
		log.Printf("wails proxy ask error question=%q elapsed=%s error=%v", question, time.Since(startedAt), err)
		return nil, err
	}
	log.Printf("wails proxy ask done question=%q elapsed=%s", question, time.Since(startedAt))
	return result, nil
}

func (a *App) Logout() {
	a.mu.Lock()
	a.studentToken = ""
	a.mu.Unlock()
}

func (a *App) postJSON(path string, payload any, auth bool, out any) error {
	if a == nil {
		return errors.New("app is not configured")
	}
	client := a.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(a.ctxOrBackground(), http.MethodPost, strings.TrimRight(a.apiBaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if auth {
		a.mu.Lock()
		token := a.studentToken
		a.mu.Unlock()
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var envelope apiResponse
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || envelope.Code != 0 {
		if envelope.Message != "" {
			return errors.New(envelope.Message)
		}
		return fmt.Errorf("HTTP %d", response.StatusCode)
	}
	if out == nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}
	return json.Unmarshal(envelope.Data, out)
}

func (a *App) ctxOrBackground() context.Context {
	if a != nil && a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
