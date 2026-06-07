package submission

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"ans-b/server/internal/qaimport"
)

func TestServiceCreateTrimsAndValidatesInput(t *testing.T) {
	repo := &fakeRepository{
		createResult: &Submission{
			ID:       3,
			UserID:   7,
			Question: "食堂几点关门？",
			Answer:   "晚上 9 点。",
			Status:   StatusPending,
		},
	}
	service := NewService(repo, nil, nil)

	created, err := service.Create(context.Background(), 7, CreateInput{
		Question: "  食堂几点关门？  ",
		Answer:   "  晚上 9 点。 ",
		Category: " 餐饮 ",
		Tags:     []string{" 食堂 ", "关门", "食堂"},
		Source:   " 后勤公告 ",
		Remark:   " 仅供测试 ",
	})
	if err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if created.ID != 3 {
		t.Fatalf("expected submission id 3, got %d", created.ID)
	}
	if repo.createInput.Question != "食堂几点关门？" {
		t.Fatalf("unexpected question: %q", repo.createInput.Question)
	}
	if repo.createInput.Answer != "晚上 9 点。" {
		t.Fatalf("unexpected answer: %q", repo.createInput.Answer)
	}
	if len(repo.createInput.Tags) != 2 {
		t.Fatalf("expected deduplicated tags, got %#v", repo.createInput.Tags)
	}
}

func TestServiceCreateRejectsMissingQuestion(t *testing.T) {
	service := NewService(&fakeRepository{}, nil, nil)

	if _, err := service.Create(context.Background(), 7, CreateInput{Answer: "ok"}); err == nil {
		t.Fatal("expected missing question to fail")
	}
}

func TestServiceListForAdminRejectsInvalidStatus(t *testing.T) {
	service := NewService(&fakeRepository{}, nil, nil)

	if _, err := service.ListForAdmin(context.Background(), "unknown"); err == nil {
		t.Fatal("expected invalid status to fail")
	}
}

func TestServiceRejectRequiresPendingStatus(t *testing.T) {
	repo := &fakeRepository{
		findResult: &Submission{ID: 8, Status: StatusApproved},
	}
	service := NewService(repo, nil, nil)

	if err := service.Reject(context.Background(), 8, ReviewInput{}); err == nil {
		t.Fatal("expected reviewed submission to fail")
	}
}

func TestServiceApproveBuildsReviewedKnowledgeAndEmbeds(t *testing.T) {
	repo := &fakeRepository{
		findResult: &Submission{
			ID:       9,
			UserID:   7,
			Question: "原问题",
			Answer:   "原答案",
			Status:   StatusPending,
		},
	}
	knowledgeRepo := &fakeKnowledgeRepository{}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	service := NewService(repo, knowledgeRepo, embedder)

	err := service.Approve(context.Background(), 9, ReviewInput{
		Question:     "  修正问题  ",
		Answer:       "  修正答案  ",
		Category:     " 餐饮 ",
		Tags:         []string{" 食堂 ", "食堂", "开放"},
		Source:       " 后勤公告 ",
		Remark:       " 已核实 ",
		ReviewerNote: " 可以入库 ",
	})
	if err != nil {
		t.Fatalf("approve submission: %v", err)
	}
	if repo.approvedID != 9 {
		t.Fatalf("expected approved id 9, got %d", repo.approvedID)
	}
	if repo.approvedNote != "可以入库" {
		t.Fatalf("unexpected reviewer note: %q", repo.approvedNote)
	}
	if repo.approvedRecord.Question != "修正问题" || repo.approvedRecord.Answer != "修正答案" {
		t.Fatalf("unexpected approved record: %#v", repo.approvedRecord)
	}
	if repo.approvedRecord.SourceType != "user_submit" {
		t.Fatalf("unexpected source type: %q", repo.approvedRecord.SourceType)
	}
	if len(repo.approvedRecord.Tags) != 2 {
		t.Fatalf("expected deduplicated tags, got %#v", repo.approvedRecord.Tags)
	}
	if len(repo.approvedRecord.Chunks) != 1 || repo.approvedRecord.Chunks[0].SourceURL != "" {
		t.Fatalf("unexpected chunk source: %#v", repo.approvedRecord.Chunks)
	}
	if repo.approvedRecord.Chunks[0].Embedding != "[0.10000000,0.20000000,0.30000000]" {
		t.Fatalf("unexpected embedding literal: %q", repo.approvedRecord.Chunks[0].Embedding)
	}
	if len(embedder.texts) != 1 || embedder.texts[0] == "" {
		t.Fatalf("expected one embedded chunk text, got %#v", embedder.texts)
	}
}

func TestServiceApproveDoesNotPersistWhenEmbeddingFails(t *testing.T) {
	repo := &fakeRepository{
		findResult: &Submission{
			ID:       9,
			Question: "问题",
			Answer:   "答案",
			Status:   StatusPending,
		},
	}
	service := NewService(repo, &fakeKnowledgeRepository{}, &fakeEmbedder{err: errors.New("embed failed")})

	if err := service.Approve(context.Background(), 9, ReviewInput{}); err == nil {
		t.Fatal("expected embedding failure")
	}
	if repo.approvedID != 0 {
		t.Fatalf("expected no approval persistence, got id %d", repo.approvedID)
	}
}

type fakeRepository struct {
	createInput    RepositoryCreateInput
	createResult   *Submission
	createErr      error
	findResult     *Submission
	findErr        error
	listByUser     []Submission
	listByStatus   []Submission
	markApproved   error
	markRejected   error
	lastListUser   int64
	lastListState  string
	approvedID     int64
	approvedNote   string
	approvedRecord qaimport.KnowledgeRecord
}

func (r *fakeRepository) Create(ctx context.Context, input RepositoryCreateInput) (*Submission, error) {
	r.createInput = input
	return r.createResult, r.createErr
}

func (r *fakeRepository) FindByID(ctx context.Context, id int64) (*Submission, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.findResult == nil {
		return nil, errors.New("not found")
	}
	return r.findResult, nil
}

func (r *fakeRepository) ListByUserID(ctx context.Context, userID int64) ([]Submission, error) {
	r.lastListUser = userID
	return r.listByUser, nil
}

func (r *fakeRepository) ListByStatus(ctx context.Context, status string) ([]Submission, error) {
	r.lastListState = status
	return r.listByStatus, nil
}

func (r *fakeRepository) ApproveWithKnowledge(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time, record qaimport.KnowledgeRecord, knowledgeRepo knowledgeTxRepository) error {
	r.approvedID = id
	r.approvedNote = reviewerNote
	r.approvedRecord = record
	return r.markApproved
}

func (r *fakeRepository) MarkRejected(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error {
	return r.markRejected
}

type fakeKnowledgeRepository struct {
	record qaimport.KnowledgeRecord
	err    error
}

func (r *fakeKnowledgeRepository) InsertKnowledgeTx(ctx context.Context, tx *sql.Tx, record qaimport.KnowledgeRecord) error {
	r.record = record
	return r.err
}

type fakeEmbedder struct {
	texts     []string
	embedding []float64
	err       error
}

func (e *fakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	e.texts = append([]string(nil), texts...)
	if e.err != nil {
		return nil, e.err
	}
	if e.embedding == nil {
		e.embedding = []float64{0.1}
	}
	return [][]float64{e.embedding}, nil
}
