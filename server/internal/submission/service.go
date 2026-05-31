package submission

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"ans-b/server/internal/embedding"
	"ans-b/server/internal/qaimport"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

type Service struct {
	repository    submissionRepository
	knowledgeRepo *qaimport.PostgresRepository
	embedder      qaimport.Embedder
}

type submissionRepository interface {
	Create(ctx context.Context, input RepositoryCreateInput) (*Submission, error)
	FindByID(ctx context.Context, id int64) (*Submission, error)
	ListByUserID(ctx context.Context, userID int64) ([]Submission, error)
	ListByStatus(ctx context.Context, status string) ([]Submission, error)
	MarkApproved(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error
	MarkRejected(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error
}

type CreateInput struct {
	Question string
	Answer   string
	Category string
	Tags     []string
	Source   string
	Remark   string
}

type ReviewInput struct {
	ReviewerNote string
}

func NewService(repository submissionRepository, knowledgeRepo *qaimport.PostgresRepository, embedder qaimport.Embedder) *Service {
	return &Service{
		repository:    repository,
		knowledgeRepo: knowledgeRepo,
		embedder:      embedder,
	}
}

func (s *Service) Create(ctx context.Context, userID int64, input CreateInput) (*Submission, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("submission service is not configured")
	}

	record := CreateInput{
		Question: strings.TrimSpace(input.Question),
		Answer:   strings.TrimSpace(input.Answer),
		Category: strings.TrimSpace(input.Category),
		Tags:     cleanTags(input.Tags),
		Source:   strings.TrimSpace(input.Source),
		Remark:   strings.TrimSpace(input.Remark),
	}
	if record.Question == "" {
		return nil, errors.New("question is required")
	}
	if record.Answer == "" {
		return nil, errors.New("answer is required")
	}
	if len(record.Category) > 100 {
		return nil, errors.New("category is too long")
	}

	return s.repository.Create(ctx, RepositoryCreateInput{
		UserID:   userID,
		Question: record.Question,
		Answer:   record.Answer,
		Category: record.Category,
		Tags:     record.Tags,
		Source:   record.Source,
		Remark:   record.Remark,
	})
}

func (s *Service) ListForStudent(ctx context.Context, userID int64) ([]Submission, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("submission service is not configured")
	}
	return s.repository.ListByUserID(ctx, userID)
}

func (s *Service) ListForAdmin(ctx context.Context, status string) ([]Submission, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("submission service is not configured")
	}
	status = normalizeStatus(status)
	if status != "" && status != StatusPending && status != StatusApproved && status != StatusRejected {
		return nil, errors.New("invalid submission status")
	}
	return s.repository.ListByStatus(ctx, status)
}

func (s *Service) Approve(ctx context.Context, submissionID int64, input ReviewInput) error {
	if submissionID <= 0 {
		return errors.New("invalid submission id")
	}
	if s == nil || s.repository == nil || s.knowledgeRepo == nil || s.embedder == nil {
		return errors.New("submission service is not configured")
	}

	submission, err := s.repository.FindByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("submission not found")
		}
		return err
	}
	if submission.Status != StatusPending {
		return errors.New("submission has already been reviewed")
	}

	record, err := s.buildKnowledgeRecord(*submission)
	if err != nil {
		return err
	}
	if err := s.publishKnowledge(ctx, record); err != nil {
		return err
	}

	return s.repository.MarkApproved(ctx, submissionID, strings.TrimSpace(input.ReviewerNote), time.Now())
}

func (s *Service) Reject(ctx context.Context, submissionID int64, input ReviewInput) error {
	if submissionID <= 0 {
		return errors.New("invalid submission id")
	}
	if s == nil || s.repository == nil {
		return errors.New("submission service is not configured")
	}

	submission, err := s.repository.FindByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("submission not found")
		}
		return err
	}
	if submission.Status != StatusPending {
		return errors.New("submission has already been reviewed")
	}

	return s.repository.MarkRejected(ctx, submissionID, strings.TrimSpace(input.ReviewerNote), time.Now())
}

func (s *Service) buildKnowledgeRecord(submission Submission) (qaimport.KnowledgeRecord, error) {
	record, err := qaimport.BuildRecord(qaimport.Item{
		Question: submission.Question,
		Answer:   submission.Answer,
		Category: submission.Category,
		Tags:     submission.Tags,
		Source:   submission.Source,
		Remark:   submission.Remark,
	})
	if err != nil {
		return qaimport.KnowledgeRecord{}, err
	}
	record.SourceType = "user_submit"
	if len(record.Chunks) > 0 {
		record.Chunks[0].SourceURL = strings.TrimSpace(submission.Source)
	}
	return record, nil
}

func (s *Service) publishKnowledge(ctx context.Context, record qaimport.KnowledgeRecord) error {
	if s.embedder == nil {
		return errors.New("embedder is not configured")
	}
	if len(record.Chunks) == 0 {
		return errors.New("knowledge chunks are required")
	}

	embeddings, err := s.embedder.Embed(ctx, []string{record.Chunks[0].Text})
	if err != nil {
		return err
	}
	if len(embeddings) != 1 {
		return errors.New("embedding count mismatch")
	}

	record.Chunks[0].Embedding = embedding.VectorLiteral(embeddings[0])
	return s.knowledgeRepo.InsertKnowledge(ctx, record)
}

func normalizeStatus(status string) string {
	return strings.TrimSpace(strings.ToLower(status))
}
