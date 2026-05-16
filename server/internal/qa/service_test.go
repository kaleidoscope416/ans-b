package qa

import (
	"context"
	"testing"
)

func TestServiceAskReturnsBestVectorAnswer(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ID:       7,
				Question: "一食堂营业时间是什么？",
				Answer:   "一食堂晚餐营业至 20:00。",
				Category: "餐饮服务",
				Score:    0.72,
			},
			{
				ID:       8,
				Question: "二食堂晚上营业到几点？",
				Answer:   "二食堂营业至 21:00。",
				Category: "餐饮服务",
				Score:    0.68,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	service := NewService(repo, embedder)

	result, err := service.Ask(context.Background(), "食堂几点关门？", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if result.Answer.Answer != "一食堂晚餐营业至 20:00。" {
		t.Fatalf("unexpected answer: %#v", result.Answer)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(result.Candidates))
	}
	if len(embedder.texts) != 1 || embedder.texts[0] != "食堂几点关门？" {
		t.Fatalf("expected question to be embedded, got %#v", embedder.texts)
	}
	if repo.queryEmbedding != "[0.10000000,0.20000000,0.30000000]" {
		t.Fatalf("unexpected query embedding: %s", repo.queryEmbedding)
	}
	if repo.limit != 5 {
		t.Fatalf("expected limit 5, got %d", repo.limit)
	}
}

func TestServiceAskRejectsEmptyQuestion(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeEmbedder{})

	if _, err := service.Ask(context.Background(), "   ", 5); err == nil {
		t.Fatal("expected empty question to fail")
	}
}

type fakeEmbedder struct {
	texts     []string
	embedding []float64
}

func (e *fakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	e.texts = texts
	return [][]float64{e.embedding}, nil
}

type fakeRepository struct {
	queryEmbedding string
	limit          int
	results        []Answer
}

func (r *fakeRepository) SearchTop(ctx context.Context, queryEmbedding string, limit int) ([]Answer, error) {
	r.queryEmbedding = queryEmbedding
	r.limit = limit
	return r.results, nil
}
