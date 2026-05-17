package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ans-b/server/internal/qa"

	openai "github.com/sashabaranov/go-openai"
)

func TestOpenAICompatibleClientGenerateAnswerPostsCandidates(t *testing.T) {
	var gotAuth string
	var gotRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("expected /v1/chat/completions, got %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"choices": []map[string]any{{"index": 0, "message": map[string]string{"role": "assistant", "content": "一食堂晚餐到 20:00 结束。"}}},
		})
	}))
	defer server.Close()

	apiKey := "test-key"
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = server.URL + "/v1"
	config.HTTPClient = &http.Client{Timeout: time.Second}
	client := &OpenAICompatibleClient{
		apiKey:  apiKey,
		baseURL: config.BaseURL,
		model:   "test-model",
		client:  openai.NewClientWithConfig(config),
	}
	answer, err := client.GenerateAnswer(context.Background(), "食堂几点关门？", []qa.Answer{
		{
			ChunkID:   17,
			ItemID:    7,
			Question:  "一食堂营业时间是什么？",
			Answer:    "一食堂早餐 6:30-9:00，午餐 10:30-13:30，晚餐 16:30-20:00。",
			ChunkText: "一食堂早餐 6:30-9:00，午餐 10:30-13:30，晚餐 16:30-20:00。",
			Category:  "餐饮服务",
			Tags:      []string{"食堂", "营业时间"},
			Score:     0.72,
		},
	}, 0.45)
	if err != nil {
		t.Fatalf("generate answer: %v", err)
	}

	if answer != "一食堂晚餐到 20:00 结束。" {
		t.Fatalf("unexpected answer: %q", answer)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotRequest.Model != "test-model" {
		t.Fatalf("unexpected model: %q", gotRequest.Model)
	}
	if len(gotRequest.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(gotRequest.Messages))
	}
	systemPrompt := gotRequest.Messages[0].Content
	for _, want := range []string{"星见 遥", "暂时无法确认", "禁止编造"} {
		if !strings.Contains(systemPrompt, want) {
			t.Fatalf("expected system prompt to contain %q, got:\n%s", want, systemPrompt)
		}
	}
	userPrompt := gotRequest.Messages[1].Content
	for _, want := range []string{"食堂几点关门？", "一食堂营业时间是什么？", "命中片段：一食堂早餐 6:30-9:00", "相似度：0.7200", "命中阈值：0.4500"} {
		if !strings.Contains(userPrompt, want) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", want, userPrompt)
		}
	}
	for _, unexpected := range []string{"片段ID：17", "知识ID：7"} {
		if strings.Contains(userPrompt, unexpected) {
			t.Fatalf("expected prompt not to contain internal id %q, got:\n%s", unexpected, userPrompt)
		}
	}
}
