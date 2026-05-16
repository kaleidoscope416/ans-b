package mockqa

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadFileReadsMockQA(t *testing.T) {
	items, err := LoadFile(filepath.Join("..", "..", "..", "data", "mock_qa.json"))
	if err != nil {
		t.Fatalf("load mock qa: %v", err)
	}
	if len(items) != 12 {
		t.Fatalf("expected 12 items, got %d", len(items))
	}
	if items[0].Question == "" || items[0].Answer == "" {
		t.Fatalf("expected first item to have question and answer: %#v", items[0])
	}
}

func TestBuildChunkTextIncludesRetrievalFields(t *testing.T) {
	chunk := BuildChunkText(Item{
		Question: "学校哪里可以自助打印？",
		Answer:   "图书馆一楼大厅可以打印。",
		Category: "打印复印",
		Tags:     []string{"打印", "图书馆"},
	})

	for _, want := range []string{"问题：学校哪里可以自助打印？", "分类：打印复印", "标签：打印，图书馆", "答案：图书馆一楼大厅可以打印。"} {
		if !contains(chunk, want) {
			t.Fatalf("expected chunk to contain %q, got:\n%s", want, chunk)
		}
	}
}

func TestEmbeddingVectorLiteralHas1024Dimensions(t *testing.T) {
	literal := VectorLiteral(EmbedText("食堂几点关门？"))
	if literal[0] != '[' || literal[len(literal)-1] != ']' {
		t.Fatalf("expected pgvector literal, got %q", literal)
	}
	if countCommas(literal) != 1023 {
		t.Fatalf("expected 1024 dimensions, got literal with %d commas", countCommas(literal))
	}
}

func TestImportItemsUsesRepository(t *testing.T) {
	repo := &fakeRepository{}
	items := []Item{{
		Question: "一食堂营业时间是什么？",
		Answer:   "一食堂晚餐营业至 20:00。",
		Category: "餐饮服务",
		Tags:     []string{"食堂", "营业时间"},
	}}

	count, err := ImportItems(context.Background(), repo, items)
	if err != nil {
		t.Fatalf("import items: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	if repo.item.Question != items[0].Question {
		t.Fatalf("repository received wrong item: %#v", repo.item)
	}
	if repo.chunkText == "" || repo.embedding == "" {
		t.Fatalf("expected chunk text and embedding to be set")
	}
}

func TestPostgresTextArrayLiteralEscapesValues(t *testing.T) {
	got := PostgresTextArrayLiteral([]string{"打印", "图书馆", `a"b`, `c\d`})
	want := `{"打印","图书馆","a\"b","c\\d"}`
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
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

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && index(s, substr) >= 0)
}

func index(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func countCommas(s string) int {
	count := 0
	for _, r := range s {
		if r == ',' {
			count++
		}
	}
	return count
}
