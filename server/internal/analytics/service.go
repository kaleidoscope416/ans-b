package analytics

import (
	"context"
	"errors"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) IncrementKnowledgeAccess(ctx context.Context, itemID int64) error {
	if s == nil || s.repository == nil {
		return errors.New("analytics service is not configured")
	}
	return s.repository.IncrementKnowledgeAccess(ctx, itemID)
}
