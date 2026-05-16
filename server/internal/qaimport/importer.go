package qaimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"ans-b/server/internal/embedding"
)

type Item struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Source   string   `json:"source"`
	Remark   string   `json:"remark"`
}

type Repository interface {
	InsertKnowledge(ctx context.Context, item Item, chunkText string, embedding string) error
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

func LoadFile(path string) ([]Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func ImportFile(ctx context.Context, repo Repository, embedder Embedder, path string) (int, error) {
	items, err := LoadFile(path)
	if err != nil {
		return 0, fmt.Errorf("load qa file: %w", err)
	}
	return ImportItems(ctx, repo, embedder, items)
}

func ImportItems(ctx context.Context, repo Repository, embedder Embedder, items []Item) (int, error) {
	if repo == nil {
		return 0, errors.New("repository is required")
	}
	if embedder == nil {
		return 0, errors.New("embedder is required")
	}

	chunkTexts := make([]string, 0, len(items))
	for _, item := range items {
		if err := validateItem(item); err != nil {
			return 0, err
		}
		chunkTexts = append(chunkTexts, BuildChunkText(item))
	}

	embeddings, err := embedder.Embed(ctx, chunkTexts)
	if err != nil {
		return 0, fmt.Errorf("embed qa chunks: %w", err)
	}
	if len(embeddings) != len(items) {
		return 0, fmt.Errorf("embedding count mismatch: got %d, want %d", len(embeddings), len(items))
	}

	count := 0
	for i, item := range items {
		if err := repo.InsertKnowledge(ctx, item, chunkTexts[i], embedding.VectorLiteral(embeddings[i])); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func BuildChunkText(item Item) string {
	parts := []string{"问题：" + strings.TrimSpace(item.Question)}
	if strings.TrimSpace(item.Category) != "" {
		parts = append(parts, "分类："+strings.TrimSpace(item.Category))
	}
	if tags := cleanTags(item.Tags); len(tags) > 0 {
		parts = append(parts, "标签："+strings.Join(tags, "，"))
	}
	parts = append(parts, "答案："+strings.TrimSpace(item.Answer))
	return strings.Join(parts, "\n")
}

func validateItem(item Item) error {
	if strings.TrimSpace(item.Question) == "" {
		return errors.New("question is required")
	}
	if strings.TrimSpace(item.Answer) == "" {
		return errors.New("answer is required")
	}
	return nil
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
