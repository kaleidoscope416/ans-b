package qa

import (
	"context"
	"testing"
)

func TestServiceAskReturnsBestVectorAnswer(t *testing.T) {
	repo := &fakeRepository{
		result: Answer{
			ID:       7,
			Question: "一食堂营业时间是什么？",
			Answer:   "一食堂晚餐营业至 20:00。",
			Category: "餐饮服务",
			Score:    0.72,
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	service := NewService(repo, embedder)

	answer, err := service.Ask(context.Background(), "食堂几点关门？")
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if answer.Answer != "一食堂晚餐营业至 20:00。" {
		t.Fatalf("unexpected answer: %#v", answer)
	}
	if len(embedder.texts) != 1 || embedder.texts[0] != "食堂几点关门？" {
		t.Fatalf("expected question to be embedded, got %#v", embedder.texts)
	}
	if repo.queryEmbedding != "[0.10000000,0.20000000,0.30000000]" {
		t.Fatalf("unexpected query embedding: %s", repo.queryEmbedding)
	}
}

func TestServiceAskRejectsEmptyQuestion(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeEmbedder{})

	if _, err := service.Ask(context.Background(), "   "); err == nil {
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
	result         Answer
}

func (r *fakeRepository) SearchBest(ctx context.Context, queryEmbedding string) (*Answer, error) {
	r.queryEmbedding = queryEmbedding
	return &r.result, nil
}
