package qaimport

import (
	"context"
	"path/filepath"
	"strings"
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
		Title:    "一食堂营业时间",
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
	if repo.record.Title != "一食堂营业时间" {
		t.Fatalf("expected title to be preserved, got %q", repo.record.Title)
	}
	if len(repo.record.Chunks) != 1 {
		t.Fatalf("expected one chunk, got %#v", repo.record.Chunks)
	}
	if len(embedder.texts) != 1 || embedder.texts[0] != repo.record.Chunks[0].Text {
		t.Fatalf("expected importer to embed chunk text, texts=%#v chunk=%q", embedder.texts, repo.record.Chunks[0].Text)
	}
	if repo.record.Chunks[0].Embedding != "[0.10000000,0.20000000,0.30000000]" {
		t.Fatalf("unexpected embedding literal: %s", repo.record.Chunks[0].Embedding)
	}
	for _, want := range []string{"问题：一食堂营业时间是什么？", "分类：餐饮服务", "标签：食堂，营业时间", "答案：一食堂晚餐营业至 20:00。"} {
		if !strings.Contains(repo.record.Chunks[0].Text, want) {
			t.Fatalf("expected chunk to contain %q, got:\n%s", want, repo.record.Chunks[0].Text)
		}
	}
}

func TestBuildRecordDefaultsTitleToQuestion(t *testing.T) {
	record, err := BuildRecord(Item{
		Question: "在哪里打印？",
		Answer:   "教学楼一楼可以自助打印。",
	})
	if err != nil {
		t.Fatalf("build record: %v", err)
	}
	if record.Title != "在哪里打印？" {
		t.Fatalf("expected title to default to question, got %q", record.Title)
	}
	if len(record.Chunks) != 1 || !strings.Contains(record.Chunks[0].Text, "答案：教学楼一楼可以自助打印。") {
		t.Fatalf("unexpected chunks: %#v", record.Chunks)
	}
}

func TestFirstChunkSourceURLReturnsFirstNonEmptyURL(t *testing.T) {
	sourceURL := firstChunkSourceURL(KnowledgeRecord{
		Chunks: []KnowledgeChunk{
			{Text: "无来源"},
			{Text: "官网片段", SourceURL: " https://www.lib.scut.edu.cn/open-hours "},
			{Text: "另一个片段", SourceURL: "https://www.lib.scut.edu.cn/other"},
		},
	})
	if sourceURL != "https://www.lib.scut.edu.cn/open-hours" {
		t.Fatalf("unexpected source url: %q", sourceURL)
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
	record KnowledgeRecord
}

func (r *fakeRepository) InsertKnowledge(ctx context.Context, record KnowledgeRecord) error {
	r.record = record
	return nil
}
