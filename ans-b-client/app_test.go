package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginStudentStoresToken(t *testing.T) {
	app := NewApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/student/login" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var request map[string]string
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request["username"] != "alice" || request["password"] != "secret123" {
			t.Fatalf("unexpected login body: %#v", request)
		}
		writeJSON(t, w, map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"token": "student-token",
				"user": map[string]any{
					"id":       7,
					"username": "alice",
					"nickname": "小爱",
					"role":     "student",
				},
			},
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	result, err := app.LoginStudent(" alice ", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Token != "student-token" {
		t.Fatalf("expected token, got %#v", result.Token)
	}
	if result.User.Nickname != "小爱" {
		t.Fatalf("expected nickname, got %#v", result.User.Nickname)
	}
	if app.getStudentToken() != "student-token" {
		t.Fatalf("expected stored token, got %#v", app.getStudentToken())
	}
}

func TestRegisterStudentRegistersThenLogsIn(t *testing.T) {
	app := NewApp()
	paths := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v1/users/register":
			var request map[string]string
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode register request: %v", err)
			}
			if request["username"] != "bob" || request["nickname"] != "小博" {
				t.Fatalf("unexpected register body: %#v", request)
			}
			writeJSON(t, w, map[string]any{
				"code":    0,
				"message": "success",
				"data":    map[string]any{"username": "bob"},
			})
		case "/api/v1/auth/student/login":
			writeJSON(t, w, map[string]any{
				"code":    0,
				"message": "success",
				"data": map[string]any{
					"token": "new-token",
					"user":  map[string]any{"username": "bob", "nickname": "小博"},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	result, err := app.RegisterStudent("bob", "secret123", "小博")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if result.Token != "new-token" {
		t.Fatalf("expected login token, got %#v", result.Token)
	}
	if app.getStudentToken() != "new-token" {
		t.Fatalf("expected stored token, got %#v", app.getStudentToken())
	}
	if len(paths) != 2 || paths[0] != "/api/v1/users/register" || paths[1] != "/api/v1/auth/student/login" {
		t.Fatalf("unexpected request order: %#v", paths)
	}
}

func TestRegisterStudentReportsAutoLoginFailure(t *testing.T) {
	app := NewApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/users/register":
			writeJSON(t, w, map[string]any{
				"code":    0,
				"message": "success",
				"data":    map[string]any{"username": "carol"},
			})
		case "/api/v1/auth/student/login":
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(t, w, map[string]any{
				"code":    50000,
				"message": "JWT_SECRET is required",
				"data":    nil,
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	_, err := app.RegisterStudent("carol", "secret123", "小可")
	if err == nil || err.Error() != "注册成功，但自动登录失败：JWT_SECRET is required" {
		t.Fatalf("expected auto-login failure message, got %v", err)
	}
}

func TestGetCurrentUserUsesStoredBearerToken(t *testing.T) {
	app := NewApp()
	app.setStudentToken("student-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/me" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer student-token" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		writeJSON(t, w, map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"id":       7,
				"username": "alice",
				"nickname": "小爱",
			},
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	profile, err := app.GetCurrentUser()
	if err != nil {
		t.Fatalf("get current user: %v", err)
	}
	if profile.Username != "alice" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestAskQuestionUsesStoredBearerToken(t *testing.T) {
	app := NewApp()
	app.setStudentToken("student-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/qa/ask" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer student-token" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode ask request: %v", err)
		}
		if request["question"] != "食堂几点关门？" || request["limit"].(float64) != 5 {
			t.Fatalf("unexpected ask body: %#v", request)
		}
		writeJSON(t, w, map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"answered":  true,
				"ai_answer": "二食堂晚餐到 21:00。",
			},
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	result, err := app.AskQuestion(" 食堂几点关门？ ", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if result["ai_answer"] != "二食堂晚餐到 21:00。" {
		t.Fatalf("unexpected ask result: %#v", result)
	}
}

func TestCreateSubmissionCleansPayloadAndUsesBearerToken(t *testing.T) {
	app := NewApp()
	app.setStudentToken("student-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/submissions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer student-token" {
			t.Fatalf("unexpected authorization: %q", got)
		}

		var request SubmissionInput
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode submission request: %v", err)
		}
		if request.Question != "图书馆几点关门？" || request.Answer != "22:00" {
			t.Fatalf("unexpected submission body: %#v", request)
		}
		if len(request.Tags) != 2 || request.Tags[0] != "图书馆" || request.Tags[1] != "时间" {
			t.Fatalf("unexpected tags: %#v", request.Tags)
		}

		writeJSON(t, w, map[string]any{
			"code":    0,
			"message": "success",
			"data": map[string]any{
				"id":         11,
				"question":   request.Question,
				"answer":     request.Answer,
				"tags":       request.Tags,
				"status":     "pending",
				"created_at": time.Now().Format(time.RFC3339),
			},
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	created, err := app.CreateSubmission(SubmissionInput{
		Question: " 图书馆几点关门？ ",
		Answer:   " 22:00 ",
		Tags:     []string{" 图书馆 ", "", "时间"},
	})
	if err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if created.ID != 11 || created.Status != "pending" {
		t.Fatalf("unexpected created submission: %#v", created)
	}
}

func TestListMySubmissionsUsesStoredBearerToken(t *testing.T) {
	app := NewApp()
	app.setStudentToken("student-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/submissions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer student-token" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		writeJSON(t, w, map[string]any{
			"code":    0,
			"message": "success",
			"data": []map[string]any{
				{
					"id":         3,
					"question":   "校园卡可以补办吗？",
					"answer":     "可以去服务大厅办理。",
					"status":     "pending",
					"created_at": time.Now().Format(time.RFC3339),
				},
			},
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	submissions, err := app.ListMySubmissions()
	if err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	if len(submissions) != 1 || submissions[0].ID != 3 {
		t.Fatalf("unexpected submissions: %#v", submissions)
	}
}

func TestLogoutCallsBackendAndClearsToken(t *testing.T) {
	app := NewApp()
	app.setStudentToken("student-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/logout" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer student-token" {
			t.Fatalf("unexpected authorization: %q", got)
		}
		writeJSON(t, w, map[string]any{
			"code":    0,
			"message": "success",
			"data":    nil,
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	if err := app.Logout(); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if app.getStudentToken() != "" {
		t.Fatalf("expected token to be cleared, got %q", app.getStudentToken())
	}
}

func TestAuthExpiryClearsStoredToken(t *testing.T) {
	app := NewApp()
	app.setStudentToken("student-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSON(t, w, map[string]any{
			"code":    40001,
			"message": "login session expired",
			"data":    nil,
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	_, err := app.ListMySubmissions()
	if err == nil || err.Error() != authExpiredText {
		t.Fatalf("expected auth expired error, got %v", err)
	}
	if app.getStudentToken() != "" {
		t.Fatalf("expected token to be cleared, got %q", app.getStudentToken())
	}
}

func TestGetHotQuestionsStatusReturnsGracefulTodoState(t *testing.T) {
	app := NewApp()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/analytics/hot-questions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "limit=6") {
			t.Fatalf("unexpected query: %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusNotImplemented)
		writeJSON(t, w, map[string]any{
			"code":    "TODO",
			"message": "TODO: implement hot question analytics server logic",
		})
	}))
	defer server.Close()
	app.apiBaseURL = server.URL

	status, err := app.GetHotQuestionsStatus(6)
	if err != nil {
		t.Fatalf("hot questions status: %v", err)
	}
	if status.Available {
		t.Fatalf("expected unavailable status, got %#v", status)
	}
	if !strings.Contains(status.Message, "hot question analytics") {
		t.Fatalf("unexpected status message: %#v", status)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}
