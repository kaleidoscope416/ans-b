package submission

import (
	"context"
	"errors"
	"testing"
	"time"
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

type fakeRepository struct {
	createInput   RepositoryCreateInput
	createResult  *Submission
	createErr     error
	findResult    *Submission
	findErr       error
	listByUser    []Submission
	listByStatus  []Submission
	markApproved  error
	markRejected  error
	lastListUser  int64
	lastListState string
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

func (r *fakeRepository) MarkApproved(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error {
	return r.markApproved
}

func (r *fakeRepository) MarkRejected(ctx context.Context, id int64, reviewerNote string, reviewedAt time.Time) error {
	return r.markRejected
}
