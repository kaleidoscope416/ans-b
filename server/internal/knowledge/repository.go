package knowledge

import (
	"context"
	"database/sql"

	"ans-b/server/internal/qaimport"
)

type Repository struct {
	inner *qaimport.PostgresRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{inner: qaimport.NewPostgresRepository(db)}
}

func (r *Repository) InsertKnowledge(ctx context.Context, record qaimport.KnowledgeRecord) error {
	if r == nil || r.inner == nil {
		return sql.ErrConnDone
	}
	return r.inner.InsertKnowledge(ctx, record)
}
