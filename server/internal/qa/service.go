package qa

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ans-b/server/internal/embedding"
)

type SearchRepository interface {
	SearchTop(ctx context.Context, queryEmbedding string, limit int) ([]Answer, error)
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

type Service struct {
	repository SearchRepository
	embedder   Embedder
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
	Answer     Answer   `json:"answer"`
	Candidates []Answer `json:"candidates"`
}

func NewService(repository SearchRepository, embedder Embedder) *Service {
	return &Service{repository: repository, embedder: embedder}
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
	return &AskResult{
		Answer:     candidates[0],
		Candidates: candidates,
	}, nil
}
