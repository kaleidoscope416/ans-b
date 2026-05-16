package qaimport

import (
	"context"
	"database/sql"
	"strings"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InsertKnowledge(ctx context.Context, item Item, chunkText string, embedding string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var itemID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM knowledge_items
		WHERE question = $1
		ORDER BY id
		LIMIT 1
	`, item.Question).Scan(&itemID)
	if err == sql.ErrNoRows {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO knowledge_items (
				title, question, answer, category, tags, source_type, status
			)
			VALUES ($1, $2, $3, $4, $5::text[], 'faq', 'approved')
			RETURNING id
		`, item.Question, item.Question, item.Answer, nullIfEmpty(item.Category), PostgresTextArrayLiteral(item.Tags)).Scan(&itemID)
	} else if err == nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE knowledge_items
			SET title = $1,
			    answer = $2,
			    category = $3,
			    tags = $4::text[],
			    source_type = 'faq',
			    status = 'approved',
			    updated_at = now()
			WHERE id = $5
		`, item.Question, item.Answer, nullIfEmpty(item.Category), PostgresTextArrayLiteral(item.Tags), itemID)
	}
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM knowledge_chunks WHERE item_id = $1`, itemID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO knowledge_chunks (item_id, chunk_text, embedding)
		VALUES ($1, $2, $3::vector)
	`, itemID, chunkText, embedding); err != nil {
		return err
	}

	err = tx.Commit()
	return err
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

func nullIfEmpty(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{String: value, Valid: value != ""}
}
