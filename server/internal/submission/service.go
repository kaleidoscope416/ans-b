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
	knowledgeRepo knowledgeTxRepository
	embedder      qaimport.Embedder
}

type submissionRepository interface {
	Create(ctx context.Context, input RepositoryCreateInput) (*Submission, error)
	FindByID(ctx context.Context, id int64) (*Submission, error)
	ListByUserID(ctx context.Context, userID int64) ([]Submission, error)
	ListByStatus(ctx context.Context, status string) ([]Submission, error)
	ApproveWithKnowledge(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time, record qaimport.KnowledgeRecord, knowledgeRepo knowledgeTxRepository) error
	MarkRejected(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error
}

type knowledgeTxRepository interface {
	InsertKnowledgeTx(ctx context.Context, tx *sql.Tx, record qaimport.KnowledgeRecord) error
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
	Question     string
	Answer       string
	Category     string
	Tags         []string
	Source       string
	Remark       string
	ReviewerNote string
}

func NewService(repository submissionRepository, knowledgeRepo knowledgeTxRepository, embedder qaimport.Embedder) *Service {
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

	reviewedSubmission, err := reviewedKnowledgeSubmission(*submission, input)
	if err != nil {
		return err
	}
	record, err := s.buildKnowledgeRecord(reviewedSubmission)
	if err != nil {
		return err
	}
	record, err = s.embedKnowledge(ctx, record)
	if err != nil {
		return err
	}

	return s.repository.ApproveWithKnowledge(ctx, submissionID, strings.TrimSpace(input.ReviewerNote), time.Now(), record, s.knowledgeRepo)
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
		record.Chunks[0].SourceURL = sourceURL(submission.Source)
	}
	return record, nil
}

func (s *Service) embedKnowledge(ctx context.Context, record qaimport.KnowledgeRecord) (qaimport.KnowledgeRecord, error) {
	if s.embedder == nil {
		return qaimport.KnowledgeRecord{}, errors.New("embedder is not configured")
	}
	if len(record.Chunks) == 0 {
		return qaimport.KnowledgeRecord{}, errors.New("knowledge chunks are required")
	}

	embeddings, err := s.embedder.Embed(ctx, []string{record.Chunks[0].Text})
	if err != nil {
		return qaimport.KnowledgeRecord{}, err
	}
	if len(embeddings) != 1 {
		return qaimport.KnowledgeRecord{}, errors.New("embedding count mismatch")
	}

	record.Chunks[0].Embedding = embedding.VectorLiteral(embeddings[0])
	return record, nil
}

func reviewedKnowledgeSubmission(submission Submission, input ReviewInput) (Submission, error) {
	reviewed := submission
	if reviewInputHasKnowledge(input) {
		if question := strings.TrimSpace(input.Question); question != "" {
			reviewed.Question = question
		}
		if answer := strings.TrimSpace(input.Answer); answer != "" {
			reviewed.Answer = answer
		}
		reviewed.Category = strings.TrimSpace(input.Category)
		reviewed.Tags = cleanTags(input.Tags)
		reviewed.Source = strings.TrimSpace(input.Source)
		reviewed.Remark = strings.TrimSpace(input.Remark)
	}
	if strings.TrimSpace(reviewed.Question) == "" {
		return Submission{}, errors.New("question is required")
	}
	if strings.TrimSpace(reviewed.Answer) == "" {
		return Submission{}, errors.New("answer is required")
	}
	if len(strings.TrimSpace(reviewed.Category)) > 100 {
		return Submission{}, errors.New("category is too long")
	}
	return reviewed, nil
}

func reviewInputHasKnowledge(input ReviewInput) bool {
	return strings.TrimSpace(input.Question) != "" ||
		strings.TrimSpace(input.Answer) != "" ||
		strings.TrimSpace(input.Category) != "" ||
		len(input.Tags) > 0 ||
		strings.TrimSpace(input.Source) != "" ||
		strings.TrimSpace(input.Remark) != ""
}

func sourceURL(source string) string {
	source = strings.TrimSpace(source)
	if strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return source
	}
	return ""
}

func normalizeStatus(status string) string {
	return strings.TrimSpace(strings.ToLower(status))
}
