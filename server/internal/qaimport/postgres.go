package qaimport

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const defaultChunkInsertBatchSize = 500

type PostgresRepository struct {
	db *sql.DB
}

type BatchOptions struct {
	ChunkBatchSize int
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InsertKnowledge(ctx context.Context, record KnowledgeRecord) error {
	_, err := r.UpsertKnowledgeBatch(ctx, []KnowledgeRecord{record}, BatchOptions{})
	return err
}

func (r *PostgresRepository) InsertKnowledgeTx(ctx context.Context, tx *sql.Tx, record KnowledgeRecord) error {
	_, err := r.UpsertKnowledgeBatchTx(ctx, tx, []KnowledgeRecord{record}, BatchOptions{})
	return err
}

func (r *PostgresRepository) UpsertKnowledgeBatch(ctx context.Context, records []KnowledgeRecord, options BatchOptions) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}
	options = normalizeBatchOptions(options)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	itemIDs := make([]int64, 0, len(records))
	chunkRows := make([]chunkInsertRow, 0)
	for _, record := range records {
		itemID, err := upsertKnowledgeItem(ctx, tx, record)
		if err != nil {
			return 0, err
		}
		itemIDs = append(itemIDs, itemID)
		for _, chunk := range record.Chunks {
			chunkRows = append(chunkRows, chunkInsertRow{
				itemID: itemID,
				chunk:  chunk,
			})
		}
	}

	if err := deleteChunksForItems(ctx, tx, itemIDs); err != nil {
		return 0, err
	}
	if err := insertChunkRows(ctx, tx, chunkRows, options.ChunkBatchSize); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	committed = true
	return len(records), nil
}

func (r *PostgresRepository) UpsertKnowledgeBatchTx(ctx context.Context, tx *sql.Tx, records []KnowledgeRecord, options BatchOptions) (int, error) {
	if tx == nil {
		return 0, fmt.Errorf("transaction is required")
	}
	if len(records) == 0 {
		return 0, nil
	}
	options = normalizeBatchOptions(options)

	itemIDs := make([]int64, 0, len(records))
	chunkRows := make([]chunkInsertRow, 0)
	for _, record := range records {
		itemID, err := upsertKnowledgeItem(ctx, tx, record)
		if err != nil {
			return 0, err
		}
		itemIDs = append(itemIDs, itemID)
		for _, chunk := range record.Chunks {
			chunkRows = append(chunkRows, chunkInsertRow{
				itemID: itemID,
				chunk:  chunk,
			})
		}
	}

	if err := deleteChunksForItems(ctx, tx, itemIDs); err != nil {
		return 0, err
	}
	if err := insertChunkRows(ctx, tx, chunkRows, options.ChunkBatchSize); err != nil {
		return 0, err
	}
	return len(records), nil
}

func upsertKnowledgeItem(ctx context.Context, tx *sql.Tx, record KnowledgeRecord) (int64, error) {
	var itemID int64
	err := findExistingItemID(ctx, tx, record).Scan(&itemID)
	if err == sql.ErrNoRows && strings.TrimSpace(record.Question) != "" {
		err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM knowledge_items
		WHERE question = $1
		ORDER BY id
		LIMIT 1
	`, record.Question).Scan(&itemID)
	}
	if err == sql.ErrNoRows {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO knowledge_items (
				title, question, answer, category, tags, source_type, status
			)
			VALUES ($1, $2, $3, $4, $5::text[], $6, 'approved')
			RETURNING id
		`, record.Title, record.Question, record.Answer, nullIfEmpty(record.Category), PostgresTextArrayLiteral(record.Tags), sourceType(record.SourceType)).Scan(&itemID)
	} else if err == nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE knowledge_items
			SET title = $1,
			    answer = $2,
			    category = $3,
			    tags = $4::text[],
			    source_type = $5,
			    status = 'approved',
			    updated_at = now()
			WHERE id = $6
		`, record.Title, record.Answer, nullIfEmpty(record.Category), PostgresTextArrayLiteral(record.Tags), sourceType(record.SourceType), itemID)
	}
	if err != nil {
		return 0, err
	}
	return itemID, nil
}

func normalizeBatchOptions(options BatchOptions) BatchOptions {
	if options.ChunkBatchSize <= 0 {
		options.ChunkBatchSize = defaultChunkInsertBatchSize
	}
	return options
}

func deleteChunksForItems(ctx context.Context, tx *sql.Tx, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(itemIDs))
	args := make([]any, 0, len(itemIDs))
	for i, itemID := range itemIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, itemID)
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM knowledge_chunks WHERE item_id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	return err
}

type chunkInsertRow struct {
	itemID int64
	chunk  KnowledgeChunk
}

func insertChunkRows(ctx context.Context, tx *sql.Tx, rows []chunkInsertRow, batchSize int) error {
	if len(rows) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = defaultChunkInsertBatchSize
	}
	for start := 0; start < len(rows); start += batchSize {
		end := start + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		query, args := buildChunkInsertQuery(rows[start:end])
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

func buildChunkInsertQuery(rows []chunkInsertRow) (string, []any) {
	var builder strings.Builder
	builder.WriteString(`INSERT INTO knowledge_chunks (item_id, chunk_text, embedding, source_url) VALUES `)
	values := make([]string, 0, len(rows))
	args := make([]any, 0, len(rows)*4)
	for i, row := range rows {
		base := i*4 + 1
		values = append(values, fmt.Sprintf("($%d, $%d, $%d::vector, $%d)", base, base+1, base+2, base+3))
		args = append(args, row.itemID, row.chunk.Text, row.chunk.Embedding, nullIfEmpty(row.chunk.SourceURL))
	}
	builder.WriteString(strings.Join(values, ","))
	return builder.String(), args
}

type rowScanner interface {
	Scan(dest ...any) error
}

func findExistingItemID(ctx context.Context, tx *sql.Tx, record KnowledgeRecord) rowScanner {
	if sourceURL := firstChunkSourceURL(record); sourceURL != "" {
		return tx.QueryRowContext(ctx, `
			SELECT ki.id
			FROM knowledge_items ki
			JOIN knowledge_chunks kc ON kc.item_id = ki.id
			WHERE kc.source_url = $1
			ORDER BY ki.id
			LIMIT 1
		`, sourceURL)
	}
	return emptyRow{}
}

type emptyRow struct{}

func (emptyRow) Scan(dest ...any) error {
	return sql.ErrNoRows
}

func firstChunkSourceURL(record KnowledgeRecord) string {
	for _, chunk := range record.Chunks {
		if sourceURL := strings.TrimSpace(chunk.SourceURL); sourceURL != "" {
			return sourceURL
		}
	}
	return ""
}

func sourceType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "faq"
	}
	return value
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
