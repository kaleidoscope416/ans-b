package qaimport

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadFileReadsSeedQA(t *testing.T) {
	items, err := LoadFile(filepath.Join("..", "..", "..", "data", "qa_seed.json"))
	if err != nil {
		t.Fatalf("load seed qa: %v", err)
	}
	if len(items) != 12 {
		t.Fatalf("expected 12 items, got %d", len(items))
	}
	if items[0].Question == "" || items[0].Answer == "" {
		t.Fatalf("expected first item to have question and answer: %#v", items[0])
	}
}

func TestImportItemsUsesEmbedderAndRepository(t *testing.T) {
	repo := &fakeRepository{}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	items := []Item{{
		Question: "一食堂营业时间是什么？",
		Answer:   "一食堂晚餐营业至 20:00。",
		Category: "餐饮服务",
		Tags:     []string{"食堂", "营业时间"},
	}}

	count, err := ImportItems(context.Background(), repo, embedder, items)
	if err != nil {
		t.Fatalf("import items: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	if len(embedder.texts) != 1 || embedder.texts[0] != repo.chunkText {
		t.Fatalf("expected importer to embed chunk text, texts=%#v chunk=%q", embedder.texts, repo.chunkText)
	}
	if repo.embedding != "[0.10000000,0.20000000,0.30000000]" {
		t.Fatalf("unexpected embedding literal: %s", repo.embedding)
	}
}

type fakeEmbedder struct {
	texts     []string
	embedding []float64
}

func (e *fakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	e.texts = texts
	return [][]float64{e.embedding}, nil
}

type fakeRepository struct {
	item      Item
	chunkText string
	embedding string
}

func (r *fakeRepository) InsertKnowledge(ctx context.Context, item Item, chunkText string, embedding string) error {
	r.item = item
	r.chunkText = chunkText
	r.embedding = embedding
	return nil
}
