package qa

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

func (r *Repository) SearchBest(ctx context.Context, queryEmbedding string) (*Answer, error) {
	if r.db == nil {
		return nil, errors.New("database is not configured")
	}

	var answer Answer
	var tags []string
	err := r.db.QueryRowContext(ctx, `
		SELECT
			ki.id,
			ki.question,
			ki.answer,
			COALESCE(ki.category, '') AS category,
			COALESCE(ki.tags, ARRAY[]::text[]) AS tags,
			1 - (kc.embedding <=> $1::vector) AS score
		FROM knowledge_chunks kc
		JOIN knowledge_items ki ON ki.id = kc.item_id
		WHERE ki.status = 'approved'
		  AND kc.embedding IS NOT NULL
		ORDER BY kc.embedding <=> $1::vector
		LIMIT 1
	`, queryEmbedding).Scan(
		&answer.ID,
		&answer.Question,
		&answer.Answer,
		&answer.Category,
		(*pqStringArray)(&tags),
		&answer.Score,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("knowledge base is empty")
	}
	if err != nil {
		return nil, err
	}
	answer.Tags = tags
	return &answer, nil
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
