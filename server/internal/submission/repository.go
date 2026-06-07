package submission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ans-b/server/internal/qaimport"
)

type Submission struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Question     string     `json:"question"`
	Answer       string     `json:"answer"`
	Category     string     `json:"category"`
	Tags         []string   `json:"tags"`
	Source       string     `json:"source"`
	Remark       string     `json:"remark"`
	Status       string     `json:"status"`
	ReviewerNote string     `json:"reviewer_note"`
	CreatedAt    time.Time  `json:"created_at"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
}

type RepositoryCreateInput struct {
	UserID   int64
	Question string
	Answer   string
	Category string
	Tags     []string
	Source   string
	Remark   string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input RepositoryCreateInput) (*Submission, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("submission repository is not configured")
	}

	var created Submission
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO user_submissions (
			user_id, question, answer, category, tags, source, remark, status
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5::text[], NULLIF($6, ''), NULLIF($7, ''), 'pending')
		RETURNING
			id,
			user_id,
			question,
			answer,
			COALESCE(category, ''),
			COALESCE(tags, ARRAY[]::text[]),
			COALESCE(source, ''),
			COALESCE(remark, ''),
			status,
			COALESCE(reviewer_note, ''),
			created_at,
			reviewed_at
	`, input.UserID, input.Question, input.Answer, input.Category, PostgresTextArrayLiteral(input.Tags), input.Source, input.Remark).Scan(
		&created.ID,
		&created.UserID,
		&created.Question,
		&created.Answer,
		&created.Category,
		(*pqStringArray)(&created.Tags),
		&created.Source,
		&created.Remark,
		&created.Status,
		&created.ReviewerNote,
		&created.CreatedAt,
		&reviewedAt,
	)
	if err != nil {
		return nil, err
	}
	created.ReviewedAt = nullTimePtr(reviewedAt)
	return &created, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Submission, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("submission repository is not configured")
	}

	var found Submission
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			user_id,
			question,
			answer,
			COALESCE(category, ''),
			COALESCE(tags, ARRAY[]::text[]),
			COALESCE(source, ''),
			COALESCE(remark, ''),
			status,
			COALESCE(reviewer_note, ''),
			created_at,
			reviewed_at
		FROM user_submissions
		WHERE id = $1
	`, id).Scan(
		&found.ID,
		&found.UserID,
		&found.Question,
		&found.Answer,
		&found.Category,
		(*pqStringArray)(&found.Tags),
		&found.Source,
		&found.Remark,
		&found.Status,
		&found.ReviewerNote,
		&found.CreatedAt,
		&reviewedAt,
	)
	if err != nil {
		return nil, err
	}
	found.ReviewedAt = nullTimePtr(reviewedAt)
	return &found, nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID int64) ([]Submission, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("submission repository is not configured")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			question,
			answer,
			COALESCE(category, ''),
			COALESCE(tags, ARRAY[]::text[]),
			COALESCE(source, ''),
			COALESCE(remark, ''),
			status,
			COALESCE(reviewer_note, ''),
			created_at,
			reviewed_at
		FROM user_submissions
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSubmissionRows(rows)
}

func (r *Repository) ListByStatus(ctx context.Context, status string) ([]Submission, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("submission repository is not configured")
	}

	query := `
		SELECT
			id,
			user_id,
			question,
			answer,
			COALESCE(category, ''),
			COALESCE(tags, ARRAY[]::text[]),
			COALESCE(source, ''),
			COALESCE(remark, ''),
			status,
			COALESCE(reviewer_note, ''),
			created_at,
			reviewed_at
		FROM user_submissions
	`
	args := make([]any, 0, 1)
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSubmissionRows(rows)
}

func (r *Repository) MarkApproved(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error {
	return r.updateReviewStatus(ctx, id, "approved", reviewerNote, reviewedAt)
}

func (r *Repository) ApproveWithKnowledge(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time, record qaimport.KnowledgeRecord, knowledgeRepo knowledgeTxRepository) error {
	if r == nil || r.db == nil {
		return errors.New("submission repository is not configured")
	}
	if knowledgeRepo == nil {
		return errors.New("knowledge repository is not configured")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	submission, err := findSubmissionByIDForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}
	if submission.Status != StatusPending {
		return errors.New("submission has already been reviewed")
	}

	if err := knowledgeRepo.InsertKnowledgeTx(ctx, tx, record); err != nil {
		return err
	}
	if err := updateReviewStatusTx(ctx, tx, id, StatusApproved, reviewerNote, reviewedAt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *Repository) MarkRejected(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("submission repository is not configured")
	}

	result, err := r.db.ExecContext(ctx, markRejectedQuery(), StatusRejected, reviewerNote, reviewedAt, id, StatusPending)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("submission has already been reviewed")
	}
	return nil
}

func markRejectedQuery() string {
	return `
		UPDATE user_submissions
		SET status = $1,
		    reviewer_note = NULLIF($2, ''),
		    reviewed_at = $3
		WHERE id = $4
		  AND status = $5
	`
}

func (r *Repository) updateReviewStatus(ctx context.Context, id int64, status, reviewerNote string, reviewedAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("submission repository is not configured")
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE user_submissions
		SET status = $1,
		    reviewer_note = NULLIF($2, ''),
		    reviewed_at = $3
		WHERE id = $4
	`, status, reviewerNote, reviewedAt, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func findSubmissionByIDForUpdate(ctx context.Context, tx *sql.Tx, id int64) (*Submission, error) {
	var found Submission
	var reviewedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			user_id,
			question,
			answer,
			COALESCE(category, ''),
			COALESCE(tags, ARRAY[]::text[]),
			COALESCE(source, ''),
			COALESCE(remark, ''),
			status,
			COALESCE(reviewer_note, ''),
			created_at,
			reviewed_at
		FROM user_submissions
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(
		&found.ID,
		&found.UserID,
		&found.Question,
		&found.Answer,
		&found.Category,
		(*pqStringArray)(&found.Tags),
		&found.Source,
		&found.Remark,
		&found.Status,
		&found.ReviewerNote,
		&found.CreatedAt,
		&reviewedAt,
	)
	if err != nil {
		return nil, err
	}
	found.ReviewedAt = nullTimePtr(reviewedAt)
	return &found, nil
}

func updateReviewStatusTx(ctx context.Context, tx *sql.Tx, id int64, status, reviewerNote string, reviewedAt time.Time) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE user_submissions
		SET status = $1,
		    reviewer_note = NULLIF($2, ''),
		    reviewed_at = $3
		WHERE id = $4
	`, status, reviewerNote, reviewedAt, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanSubmissionRows(rows *sql.Rows) ([]Submission, error) {
	submissions := make([]Submission, 0)
	for rows.Next() {
		var submission Submission
		var reviewedAt sql.NullTime
		if err := rows.Scan(
			&submission.ID,
			&submission.UserID,
			&submission.Question,
			&submission.Answer,
			&submission.Category,
			(*pqStringArray)(&submission.Tags),
			&submission.Source,
			&submission.Remark,
			&submission.Status,
			&submission.ReviewerNote,
			&submission.CreatedAt,
			&reviewedAt,
		); err != nil {
			return nil, err
		}
		submission.ReviewedAt = nullTimePtr(reviewedAt)
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

type pqStringArray []string

func (a *pqStringArray) Scan(src any) error {
	switch value := src.(type) {
	case nil:
		*a = nil
		return nil
	case string:
		return parseTextArray(value, a)
	case []byte:
		return parseTextArray(string(value), a)
	default:
		return errors.New("unsupported text array type")
	}
}

func parseTextArray(value string, out *pqStringArray) error {
	if value == "{}" || value == "" {
		*out = nil
		return nil
	}
	if len(value) < 2 || value[0] != '{' || value[len(value)-1] != '}' {
		return errors.New("invalid text array")
	}

	var result []string
	var current []rune
	inQuote := false
	escaped := false
	for _, r := range value[1 : len(value)-1] {
		if escaped {
			current = append(current, r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '"' {
			inQuote = !inQuote
			continue
		}
		if r == ',' && !inQuote {
			result = append(result, string(current))
			current = nil
			continue
		}
		current = append(current, r)
	}
	result = append(result, string(current))
	*out = result
	return nil
}

func PostgresTextArrayLiteral(values []string) string {
	if len(values) == 0 {
		return "{}"
	}
	escaped := make([]string, 0, len(values))
	for _, value := range cleanTags(values) {
		value = strings.ReplaceAll(value, `\`, `\\`)
		value = strings.ReplaceAll(value, `"`, `\"`)
		escaped = append(escaped, `"`+value+`"`)
	}
	if len(escaped) == 0 {
		return "{}"
	}
	return "{" + strings.Join(escaped, ",") + "}"
}

func cleanTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		cleaned = append(cleaned, tag)
	}
	return cleaned
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}
