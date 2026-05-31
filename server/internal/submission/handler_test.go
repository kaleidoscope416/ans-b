package submission

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ans-b/server/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestHandlerCreateRequiresStudentClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewHandler(NewService(&fakeRepository{}, nil, nil))
	engine.POST("/submissions", handler.Create)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(`{"question":"q","answer":"a"}`))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerCreateReturnsCreatedSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	repo := &fakeRepository{
		createResult: &Submission{
			ID:       4,
			UserID:   7,
			Question: "食堂几点关门？",
			Answer:   "晚上 9 点。",
			Status:   StatusPending,
		},
	}
	handler := NewHandler(NewService(repo, nil, nil))
	engine.POST("/submissions", injectClaims(auth.Claims{UserID: 7, Role: auth.RoleStudent}), handler.Create)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(`{"question":"食堂几点关门？","answer":"晚上9点。","tags":["食堂"]}`))
	request.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response struct {
		Code int `json:"code"`
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Code != 0 || response.Data.ID != 4 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func injectClaims(claims auth.Claims) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth.claims", &claims)
		c.Next()
	}
}
