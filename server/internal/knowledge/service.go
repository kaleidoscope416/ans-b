package knowledge

import (
	"context"
	"errors"
	"strings"

	"ans-b/server/internal/qaimport"
)

type Service struct {
	repository qaimport.Repository
	embedder   qaimport.Embedder
}

type CreateInput struct {
	Question string
	Answer   string
	Category string
	Tags     []string
	Source   string
	Remark   string
}

func NewService(repository qaimport.Repository, embedder qaimport.Embedder) *Service {
	return &Service{repository: repository, embedder: embedder}
}

func (s *Service) Create(ctx context.Context, input CreateInput) error {
	if s.repository == nil {
		return errors.New("knowledge repository is not configured")
	}
	if s.embedder == nil {
		return errors.New("embedder is not configured")
	}

	item := qaimport.Item{
		Question: strings.TrimSpace(input.Question),
		Answer:   strings.TrimSpace(input.Answer),
		Category: strings.TrimSpace(input.Category),
		Tags:     input.Tags,
		Source:   strings.TrimSpace(input.Source),
		Remark:   strings.TrimSpace(input.Remark),
	}
	_, err := qaimport.ImportItems(ctx, s.repository, s.embedder, []qaimport.Item{item})
	return err
}
