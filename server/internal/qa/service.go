package qa

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"ans-b/server/internal/embedding"
)

type SearchRepository interface {
	SearchTop(ctx context.Context, queryEmbedding string, limit int) ([]Answer, error)
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

type AnswerGenerator interface {
	GenerateAnswer(ctx context.Context, question string, candidates []Answer, minScore float64) (string, error)
}

type Service struct {
	repository SearchRepository
	embedder   Embedder
	generator  AnswerGenerator
	minScore   float64
}

type Answer struct {
	ID       int64    `json:"id"`
	Question string   `json:"matched_question"`
	Answer   string   `json:"answer"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Score    float64  `json:"score"`
}

type AskResult struct {
	Answered   bool     `json:"answered"`
	Answer     *Answer  `json:"answer"`
	Candidates []Answer `json:"candidates"`
	MinScore   float64  `json:"min_score"`
	AIAnswer   string   `json:"ai_answer"`
	AIEnabled  bool     `json:"ai_enabled"`
	AIError    string   `json:"ai_error"`
}

func NewService(repository SearchRepository, embedder Embedder, generators ...AnswerGenerator) *Service {
	service := &Service{
		repository: repository,
		embedder:   embedder,
		minScore:   defaultMinScore(),
	}
	if len(generators) > 0 {
		service.generator = generators[0]
	}
	return service
}

func (s *Service) Ask(ctx context.Context, question string, limit int) (*AskResult, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("question is required")
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}

	if s.embedder == nil {
		return nil, errors.New("embedder is not configured")
	}
	embeddings, err := s.embedder.Embed(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("embed question: %w", err)
	}
	if len(embeddings) != 1 {
		return nil, fmt.Errorf("embedding count mismatch: got %d, want 1", len(embeddings))
	}

	queryEmbedding := embedding.VectorLiteral(embeddings[0])
	candidates, err := s.repository.SearchTop(ctx, queryEmbedding, limit)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, errors.New("no relevant answer found")
	}
	result := &AskResult{
		Answered:   candidates[0].Score >= s.minScore,
		Candidates: candidates,
		MinScore:   s.minScore,
	}
	if result.Answered {
		result.Answer = &candidates[0]
	}
	if result.Answered && s.generator != nil {
		result.AIEnabled = true
		aiAnswer, err := s.generator.GenerateAnswer(ctx, question, candidates, s.minScore)
		if err != nil {
			result.AIError = err.Error()
			return result, nil
		}
		result.AIAnswer = strings.TrimSpace(aiAnswer)
	}
	return result, nil
}

func defaultMinScore() float64 {
	value := strings.TrimSpace(os.Getenv("QA_MIN_SCORE"))
	if value == "" {
		return 0.45
	}
	score, err := strconv.ParseFloat(value, 64)
	if err != nil || score < 0 || score > 1 {
		return 0.45
	}
	return score
}
