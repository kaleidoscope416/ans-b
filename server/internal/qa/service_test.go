package qa

import (
	"context"
	"errors"
	"testing"
)

func TestServiceAskReturnsBestVectorAnswer(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ID:        7,
				ChunkID:   17,
				ItemID:    7,
				Question:  "一食堂营业时间是什么？",
				Answer:    "一食堂晚餐营业至 20:00。",
				ChunkText: "一食堂晚餐营业至 20:00。",
				Category:  "餐饮服务",
				Score:     0.72,
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
	if !result.Answered {
		t.Fatal("expected result to be answered")
	}
	if result.Answer == nil || result.Answer.Answer != "一食堂晚餐营业至 20:00。" {
		t.Fatalf("unexpected answer: %#v", result.Answer)
	}
	if result.Answer.ChunkID != 17 || result.Answer.ItemID != 7 {
		t.Fatalf("expected best answer to include chunk identity, got %#v", result.Answer)
	}
	if result.Answer.ChunkText != "一食堂晚餐营业至 20:00。" {
		t.Fatalf("expected best answer to include chunk text, got %q", result.Answer.ChunkText)
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

func TestServiceAskDoesNotAnswerBelowMinScore(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ID:       9,
				Question: "校车时刻表在哪里看？",
				Answer:   "校车时刻表可以在后勤服务栏目查看。",
				Score:    0.24,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	service := NewService(repo, embedder)

	result, err := service.Ask(context.Background(), "asdkjhasdkjh", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if result.Answered {
		t.Fatal("expected low score result to be unanswered")
	}
	if result.Answer != nil {
		t.Fatalf("expected no answer, got %#v", result.Answer)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("expected candidates to remain visible, got %d", len(result.Candidates))
	}
	if result.MinScore != 0.45 {
		t.Fatalf("expected min score 0.45, got %f", result.MinScore)
	}
}

func TestServiceAskAddsAIAnswerWhenGeneratorSucceeds(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ID:       7,
				Question: "一食堂营业时间是什么？",
				Answer:   "一食堂晚餐营业至 20:00。",
				Category: "餐饮服务",
				Score:    0.72,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	generator := &fakeGenerator{answer: "一食堂晚餐到 20:00 结束。"}
	service := NewService(repo, embedder, generator)

	result, err := service.Ask(context.Background(), "食堂几点关门？", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if !result.AIEnabled {
		t.Fatal("expected AI to be enabled")
	}
	if result.AIAnswer != "一食堂晚餐到 20:00 结束。" {
		t.Fatalf("unexpected AI answer: %q", result.AIAnswer)
	}
	if generator.question != "食堂几点关门？" {
		t.Fatalf("expected generator to receive question, got %q", generator.question)
	}
	if len(generator.candidates) != 1 || generator.candidates[0].Score != 0.72 {
		t.Fatalf("expected generator to receive candidates, got %#v", generator.candidates)
	}
}

func TestServiceAskKeepsSearchAnswerWhenAIGenerationFails(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ID:       7,
				Question: "一食堂营业时间是什么？",
				Answer:   "一食堂晚餐营业至 20:00。",
				Score:    0.72,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	generator := &fakeGenerator{err: errors.New("upstream unavailable")}
	service := NewService(repo, embedder, generator)

	result, err := service.Ask(context.Background(), "食堂几点关门？", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if !result.AIEnabled {
		t.Fatal("expected AI to be enabled")
	}
	if result.AIAnswer != "" {
		t.Fatalf("expected empty AI answer, got %q", result.AIAnswer)
	}
	if result.AIError == "" {
		t.Fatal("expected AI error to be recorded")
	}
	if result.Answer == nil {
		t.Fatal("expected search answer to remain available")
	}
}

func TestServiceAskDoesNotCallAIWhenBelowMinScore(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ID:       9,
				Question: "校车时刻表在哪里看？",
				Answer:   "校车时刻表可以在后勤服务栏目查看。",
				Score:    0.24,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	generator := &fakeGenerator{answer: "不应调用"}
	service := NewService(repo, embedder, generator)

	result, err := service.Ask(context.Background(), "asdkjhasdkjh", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if result.Answered {
		t.Fatal("expected low score result to be unanswered")
	}
	if generator.called {
		t.Fatal("expected generator not to be called")
	}
	if result.AIEnabled {
		t.Fatal("expected AI to be disabled for unanswered result")
	}
}

func TestServiceAskIncrementsAccessWhenAnswered(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ItemID:   42,
				Question: "图书馆几点关门？",
				Answer:   "图书馆 22:00 闭馆。",
				Score:    0.81,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	recorder := &fakeAccessRecorder{}
	service := NewService(repo, embedder)
	service.SetAccessRecorder(recorder)

	result, err := service.Ask(context.Background(), "图书馆几点关门？", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if !result.Answered {
		t.Fatal("expected result to be answered")
	}
	if recorder.itemID != 42 || recorder.calls != 1 {
		t.Fatalf("expected access count for item 42 once, got item %d calls %d", recorder.itemID, recorder.calls)
	}
}

func TestServiceAskDoesNotIncrementAccessWhenBelowMinScore(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ItemID:   42,
				Question: "图书馆几点关门？",
				Answer:   "图书馆 22:00 闭馆。",
				Score:    0.2,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	recorder := &fakeAccessRecorder{}
	service := NewService(repo, embedder)
	service.SetAccessRecorder(recorder)

	result, err := service.Ask(context.Background(), "unknown", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if result.Answered {
		t.Fatal("expected result to be unanswered")
	}
	if recorder.calls != 0 {
		t.Fatalf("expected no access count increment, got %d", recorder.calls)
	}
}

func TestServiceAskIgnoresAccessRecorderError(t *testing.T) {
	repo := &fakeRepository{
		results: []Answer{
			{
				ItemID:   42,
				Question: "图书馆几点关门？",
				Answer:   "图书馆 22:00 闭馆。",
				Score:    0.81,
			},
		},
	}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	recorder := &fakeAccessRecorder{err: errors.New("database unavailable")}
	service := NewService(repo, embedder)
	service.SetAccessRecorder(recorder)

	result, err := service.Ask(context.Background(), "图书馆几点关门？", 5)
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if !result.Answered {
		t.Fatal("expected result to be answered")
	}
	if recorder.calls != 1 {
		t.Fatalf("expected recorder to be called once, got %d", recorder.calls)
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

type fakeGenerator struct {
	called     bool
	question   string
	candidates []Answer
	answer     string
	err        error
}

func (g *fakeGenerator) GenerateAnswer(ctx context.Context, question string, candidates []Answer, minScore float64) (string, error) {
	g.called = true
	g.question = question
	g.candidates = candidates
	return g.answer, g.err
}

type fakeAccessRecorder struct {
	calls  int
	itemID int64
	err    error
}

func (r *fakeAccessRecorder) IncrementKnowledgeAccess(ctx context.Context, itemID int64) error {
	r.calls++
	r.itemID = itemID
	return r.err
}
