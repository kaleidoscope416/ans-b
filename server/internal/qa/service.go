package qa

import (
	"context"
	"errors"
	"strings"

	"ans-b/server/internal/mockqa"
)

type SearchRepository interface {
	SearchBest(ctx context.Context, queryEmbedding string) (*Answer, error)
}

type Service struct {
	repository SearchRepository
}

type Answer struct {
	ID       int64    `json:"id"`
	Question string   `json:"matched_question"`
	Answer   string   `json:"answer"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Score    float64  `json:"score"`
}

func NewService(repository SearchRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Ask(ctx context.Context, question string) (*Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, errors.New("question is required")
	}

	queryEmbedding := mockqa.VectorLiteral(mockqa.EmbedText(question))
	answer, err := s.repository.SearchBest(ctx, queryEmbedding)
	if err != nil {
		return nil, err
	}
	if answer == nil {
		return nil, errors.New("no relevant answer found")
	}
	return answer, nil
}
