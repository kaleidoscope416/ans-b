package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClientEmbedPostsTextsAndReturnsEmbeddings(t *testing.T) {
	var gotTexts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/embed" {
			t.Fatalf("expected /embed, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected json content type, got %s", r.Header.Get("Content-Type"))
		}

		var req struct {
			Texts []string `json:"texts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotTexts = req.Texts

		_ = json.NewEncoder(w).Encode(map[string]any{
			"embeddings": [][]float64{{0.1, 0.2, 0.3}},
		})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	embeddings, err := client.Embed(context.Background(), []string{"食堂几点关门？"})
	if err != nil {
		t.Fatalf("embed: %v", err)
	}

	if len(gotTexts) != 1 || gotTexts[0] != "食堂几点关门？" {
		t.Fatalf("unexpected posted texts: %#v", gotTexts)
	}
	if len(embeddings) != 1 || len(embeddings[0]) != 3 {
		t.Fatalf("unexpected embeddings: %#v", embeddings)
	}
}

func TestVectorLiteralFormatsPgvectorValue(t *testing.T) {
	got := VectorLiteral([]float64{0.1, -0.25, 1})
	want := "[0.10000000,-0.25000000,1.00000000]"
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
