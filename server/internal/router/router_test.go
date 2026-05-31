package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesAddsHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	RegisterRoutes(engine)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected /healthz to return 200, got %d", recorder.Code)
	}
}

func TestRegisterRoutesHandlesCORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	RegisterRoutes(engine)

	for _, origin := range []string{
		"http://127.0.0.1:23457",
		"http://100.115.97.57:23457",
	} {
		t.Run(origin, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodOptions, "/api/v1/qa/ask", nil)
			request.Header.Set("Origin", origin)
			request.Header.Set("Access-Control-Request-Method", "POST")
			request.Header.Set("Access-Control-Request-Headers", "content-type,authorization")
			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("expected CORS preflight to return 204, got %d", recorder.Code)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("expected CORS origin header %q, got %q", origin, got)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
				t.Fatalf("expected CORS headers to allow authorization, got %q", got)
			}
		})
	}
}

func TestRegisterRoutesAddsTodoEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	RegisterRoutes(engine)

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodPost, "/api/v1/auth/student/login", http.StatusBadRequest},
		{http.MethodPost, "/api/v1/auth/admin/login", http.StatusBadRequest},
		{http.MethodPost, "/api/v1/users/register", http.StatusBadRequest},
		{http.MethodGet, "/api/v1/users/me", http.StatusUnauthorized},
		{http.MethodGet, "/api/v1/knowledge", http.StatusNotImplemented},
		{http.MethodPost, "/api/v1/knowledge", http.StatusBadRequest},
		{http.MethodPost, "/api/v1/qa/ask", http.StatusUnauthorized},
		{http.MethodGet, "/api/v1/search/candidates", http.StatusNotImplemented},
		{http.MethodPost, "/api/v1/submissions", http.StatusNotImplemented},
		{http.MethodGet, "/api/v1/submissions", http.StatusNotImplemented},
		{http.MethodGet, "/api/v1/analytics/hot-questions", http.StatusNotImplemented},
		{http.MethodPost, "/api/v1/model/embeddings", http.StatusNotImplemented},
		{http.MethodPost, "/api/v1/storage/imports", http.StatusNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)

			engine.ServeHTTP(recorder, request)

			if recorder.Code != tt.want {
				t.Fatalf("expected %s %s to return %d, got %d", tt.method, tt.path, tt.want, recorder.Code)
			}
		})
	}
}
