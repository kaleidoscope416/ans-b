package analytics

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IncrementKnowledgeAccess(ctx context.Context, itemID int64) error {
	if r == nil || r.db == nil {
		return errors.New("analytics repository is not configured")
	}
	if itemID <= 0 {
		return errors.New("knowledge item id is required")
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE knowledge_items
		SET access_count = access_count + 1,
		    last_accessed_at = now()
		WHERE id = $1
	`, itemID)
	return err
}
