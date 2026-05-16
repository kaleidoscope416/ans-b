package qa

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ans-b/server/internal/embedding"
)

type SearchRepository interface {
	SearchBest(ctx context.Context, queryEmbedding string) (*Answer, error)
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

func NewService(repository SearchRepository, embedder Embedder) *Service {
	return &Service{repository: repository, embedder: embedder}
}

func (s *Service) Ask(ctx context.Context, question string) (*Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("question is required")
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
	answer, err := s.repository.SearchBest(ctx, queryEmbedding)
	if err != nil {
		return nil, err
	}
	if answer == nil {
		return nil, errors.New("no relevant answer found")
	}
	return answer, nil
}
