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
	service := NewService(repo)

	answer, err := service.Ask(context.Background(), "食堂几点关门？")
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if answer.Answer != "一食堂晚餐营业至 20:00。" {
		t.Fatalf("unexpected answer: %#v", answer)
	}
	if repo.queryEmbedding == "" {
		t.Fatal("expected query embedding to be generated")
	}
}

func TestServiceAskRejectsEmptyQuestion(t *testing.T) {
	service := NewService(&fakeRepository{})

	if _, err := service.Ask(context.Background(), "   "); err == nil {
		t.Fatal("expected empty question to fail")
	}
}

type fakeRepository struct {
	queryEmbedding string
	result         Answer
}

func (r *fakeRepository) SearchBest(ctx context.Context, queryEmbedding string) (*Answer, error) {
	r.queryEmbedding = queryEmbedding
	return &r.result, nil
}
