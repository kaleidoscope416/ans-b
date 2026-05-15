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

func TestRegisterRoutesAddsTodoEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	RegisterRoutes(engine)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/auth/student/login"},
		{http.MethodPost, "/api/v1/auth/admin/login"},
		{http.MethodPost, "/api/v1/users/register"},
		{http.MethodGet, "/api/v1/users/me"},
		{http.MethodGet, "/api/v1/knowledge"},
		{http.MethodPost, "/api/v1/knowledge"},
		{http.MethodPost, "/api/v1/qa/ask"},
		{http.MethodGet, "/api/v1/search/candidates"},
		{http.MethodPost, "/api/v1/submissions"},
		{http.MethodGet, "/api/v1/submissions"},
		{http.MethodGet, "/api/v1/analytics/hot-questions"},
		{http.MethodPost, "/api/v1/model/embeddings"},
		{http.MethodPost, "/api/v1/storage/imports"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNotImplemented {
				t.Fatalf("expected %s %s to return 501, got %d", tt.method, tt.path, recorder.Code)
			}
		})
	}
}
