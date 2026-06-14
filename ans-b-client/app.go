package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultAPIBaseURL = "http://127.0.0.1:23456"
	authExpiredText   = "登录已过期，请重新登录"
)

// App bridges the desktop UI and the HTTP API.
type App struct {
	ctx          context.Context
	apiBaseURL   string
	httpClient   *http.Client
	mu           sync.Mutex
	studentToken string
}

type apiResponse struct {
	Code    any             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type Account struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

type UserProfile struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LoginResult struct {
	Token     string  `json:"token"`
	ExpiresIn int64   `json:"expires_in"`
	User      Account `json:"user"`
}

type SubmissionInput struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Source   string   `json:"source"`
	Remark   string   `json:"remark"`
}

type Submission struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Question     string     `json:"question"`
	Answer       string     `json:"answer"`
	Category     string     `json:"category"`
	Tags         []string   `json:"tags"`
	Source       string     `json:"source"`
	Remark       string     `json:"remark"`
	Status       string     `json:"status"`
	ReviewerNote string     `json:"reviewer_note"`
	CreatedAt    time.Time  `json:"created_at"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
}

type HotQuestionsStatus struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

// NewApp creates a new App application struct.
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

// startup is called when the app starts. The context is saved so we can call runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LoginStudent(username string, password string) (*LoginResult, error) {
	log.Printf("wails proxy login start username=%q", strings.TrimSpace(username))
	var result LoginResult
	if err := a.doJSON(http.MethodPost, "/api/v1/auth/student/login", map[string]any{
		"username": strings.TrimSpace(username),
		"password": strings.TrimSpace(password),
	}, false, &result); err != nil {
		log.Printf("wails proxy login error username=%q error=%v", strings.TrimSpace(username), err)
		return nil, err
	}
	a.setStudentToken(result.Token)
	log.Printf("wails proxy login done username=%q", strings.TrimSpace(username))
	return &result, nil
}

func (a *App) RegisterStudent(username string, password string, nickname string) (*LoginResult, error) {
	log.Printf("wails proxy register start username=%q", strings.TrimSpace(username))
	if err := a.doJSON(http.MethodPost, "/api/v1/users/register", map[string]any{
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

func (a *App) GetCurrentUser() (*UserProfile, error) {
	var profile UserProfile
	if err := a.doJSON(http.MethodGet, "/api/v1/users/me", nil, true, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
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
	if err := a.doJSON(http.MethodPost, "/api/v1/qa/ask", map[string]any{
		"question": question,
		"limit":    limit,
	}, true, &result); err != nil {
		log.Printf("wails proxy ask error question=%q elapsed=%s error=%v", question, time.Since(startedAt), err)
		return nil, err
	}
	log.Printf("wails proxy ask done question=%q elapsed=%s", question, time.Since(startedAt))
	return result, nil
}

func (a *App) CreateSubmission(input SubmissionInput) (*Submission, error) {
	payload := SubmissionInput{
		Question: strings.TrimSpace(input.Question),
		Answer:   strings.TrimSpace(input.Answer),
		Category: strings.TrimSpace(input.Category),
		Tags:     cleanTags(input.Tags),
		Source:   strings.TrimSpace(input.Source),
		Remark:   strings.TrimSpace(input.Remark),
	}

	var created Submission
	if err := a.doJSON(http.MethodPost, "/api/v1/submissions", payload, true, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (a *App) ListMySubmissions() ([]Submission, error) {
	var submissions []Submission
	if err := a.doJSON(http.MethodGet, "/api/v1/submissions", nil, true, &submissions); err != nil {
		return nil, err
	}
	if submissions == nil {
		submissions = []Submission{}
	}
	return submissions, nil
}

func (a *App) GetHotQuestionsStatus(limit int) (*HotQuestionsStatus, error) {
	if limit <= 0 {
		limit = 10
	}
	query := url.Values{}
	query.Set("limit", fmt.Sprintf("%d", limit))

	statusCode, envelope, err := a.doRequest(http.MethodGet, "/api/v1/analytics/hot-questions?"+query.Encode(), nil, false)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotImplemented || responseCodeEquals(envelope.Code, "TODO") {
		message := strings.TrimSpace(envelope.Message)
		if message == "" {
			message = "热点问题功能开发中"
		}
		return &HotQuestionsStatus{
			Available: false,
			Message:   message,
		}, nil
	}
	if apiErr := a.responseError(statusCode, envelope); apiErr != nil {
		return nil, apiErr
	}
	return &HotQuestionsStatus{
		Available: true,
		Message:   strings.TrimSpace(envelope.Message),
	}, nil
}

func (a *App) Logout() error {
	token := a.getStudentToken()
	if token == "" {
		return nil
	}
	statusCode, envelope, err := a.doRequest(http.MethodPost, "/api/v1/auth/logout", nil, true)
	if err != nil {
		return err
	}
	apiErr := a.responseError(statusCode, envelope)
	if apiErr != nil {
		return apiErr
	}
	a.clearStudentToken()
	return nil
}

func (a *App) doJSON(method string, path string, payload any, auth bool, out any) error {
	statusCode, envelope, err := a.doRequest(method, path, payload, auth)
	if err != nil {
		return err
	}
	if apiErr := a.responseError(statusCode, envelope); apiErr != nil {
		return apiErr
	}
	if out == nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil
	}
	return json.Unmarshal(envelope.Data, out)
}

func (a *App) doRequest(method string, path string, payload any, auth bool) (int, apiResponse, error) {
	if a == nil {
		return 0, apiResponse{}, errors.New("app is not configured")
	}
	client := a.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	var bodyReader *bytes.Reader
	if payload == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		body, err := json.Marshal(payload)
		if err != nil {
			return 0, apiResponse{}, err
		}
		bodyReader = bytes.NewReader(body)
	}

	request, err := http.NewRequestWithContext(a.ctxOrBackground(), method, strings.TrimRight(a.apiBaseURL, "/")+path, bodyReader)
	if err != nil {
		return 0, apiResponse{}, err
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if auth {
		token := a.getStudentToken()
		if token == "" {
			return 0, apiResponse{}, a.authExpiredError()
		}
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, apiResponse{}, err
	}
	defer response.Body.Close()

	var envelope apiResponse
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return 0, apiResponse{}, err
	}

	return response.StatusCode, envelope, nil
}

func (a *App) responseError(statusCode int, envelope apiResponse) error {
	if statusCode >= 200 && statusCode < 300 && responseCodeEquals(envelope.Code, float64(0)) {
		return nil
	}
	if statusCode == http.StatusUnauthorized || responseCodeEquals(envelope.Code, float64(40001)) {
		return a.authExpiredError()
	}
	if envelope.Message != "" {
		return errors.New(envelope.Message)
	}
	if statusCode == 0 {
		return errors.New("request failed")
	}
	return fmt.Errorf("HTTP %d", statusCode)
}

func (a *App) authExpiredError() error {
	a.clearStudentToken()
	return errors.New(authExpiredText)
}

func (a *App) setStudentToken(token string) {
	a.mu.Lock()
	a.studentToken = strings.TrimSpace(token)
	a.mu.Unlock()
}

func (a *App) getStudentToken() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return strings.TrimSpace(a.studentToken)
}

func (a *App) clearStudentToken() {
	a.mu.Lock()
	a.studentToken = ""
	a.mu.Unlock()
}

func cleanTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		cleaned = append(cleaned, tag)
	}
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func responseCodeEquals(actual any, expected any) bool {
	switch want := expected.(type) {
	case float64:
		value, ok := actual.(float64)
		return ok && value == want
	case string:
		value, ok := actual.(string)
		return ok && strings.TrimSpace(value) == want
	default:
		return false
	}
}

func (a *App) ctxOrBackground() context.Context {
	if a != nil && a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
