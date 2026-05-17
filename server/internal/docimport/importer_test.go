package docimport

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"ans-b/server/internal/qaimport"
)

func TestLoadFileReadsPages(t *testing.T) {
	pages, err := LoadFile(filepath.Join("..", "..", "..", "data", "scut_pages.json"))
	if err != nil {
		t.Fatalf("load pages: %v", err)
	}
	if len(pages) == 0 {
		t.Fatal("expected sample pages")
	}
	if pages[0].Title == "" || pages[0].Content == "" || pages[0].SourceURL == "" {
		t.Fatalf("expected first page to have title/content/source_url: %#v", pages[0])
	}
}

func TestBuildRecordRequiresTitleContentAndSourceURL(t *testing.T) {
	cases := []Page{
		{Content: "正文", SourceURL: "https://example.edu/a"},
		{Title: "标题", SourceURL: "https://example.edu/a"},
		{Title: "标题", Content: "正文"},
	}
	for _, page := range cases {
		if _, err := BuildRecord(page); err == nil {
			t.Fatalf("expected invalid page to fail: %#v", page)
		}
	}
}

func TestBuildRecordCreatesOfficialPageChunks(t *testing.T) {
	content := strings.Repeat("华南理工大学图书馆开放时间说明。", 80)
	record, err := BuildRecord(Page{
		Title:      "图书馆开放时间",
		Content:    content,
		Category:   "图书馆",
		SourceName: "华南理工大学图书馆",
		SourceURL:  "https://www.lib.scut.edu.cn/open-hours",
		Tags:       []string{"图书馆", "开放时间", "图书馆"},
	})
	if err != nil {
		t.Fatalf("build record: %v", err)
	}
	if record.Title != "图书馆开放时间" {
		t.Fatalf("unexpected title: %q", record.Title)
	}
	if record.Question != "图书馆开放时间" {
		t.Fatalf("expected question compatibility title, got %q", record.Question)
	}
	if record.Answer != content {
		t.Fatalf("expected full content in answer compatibility field")
	}
	if record.Source != "华南理工大学图书馆" {
		t.Fatalf("unexpected source: %q", record.Source)
	}
	if record.SourceType != "official_page" {
		t.Fatalf("unexpected source type: %q", record.SourceType)
	}
	if len(record.Tags) != 2 {
		t.Fatalf("expected deduplicated tags, got %#v", record.Tags)
	}
	if len(record.Chunks) < 2 {
		t.Fatalf("expected long content to split into multiple chunks, got %d", len(record.Chunks))
	}
	for _, chunk := range record.Chunks {
		if strings.TrimSpace(chunk.Text) == "" {
			t.Fatalf("expected non-empty chunk: %#v", record.Chunks)
		}
		if chunk.SourceURL != "https://www.lib.scut.edu.cn/open-hours" {
			t.Fatalf("unexpected chunk source url: %q", chunk.SourceURL)
		}
		if len([]rune(chunk.Text)) > defaultChunkSize {
			t.Fatalf("expected chunk length <= %d, got %d", defaultChunkSize, len([]rune(chunk.Text)))
		}
	}
}

func TestSplitContentUsesDefaultBGEFriendlyChunkSize(t *testing.T) {
	content := strings.Repeat("校内服务说明", 80)
	chunks := SplitContent(content)
	if len(chunks) < 2 {
		t.Fatalf("expected content to split with %d char chunks, got %d", defaultChunkSize, len(chunks))
	}
	if len([]rune(chunks[0])) != defaultChunkSize {
		t.Fatalf("expected first chunk length %d, got %d", defaultChunkSize, len([]rune(chunks[0])))
	}
	overlapStart := defaultChunkSize - defaultChunkOverlap
	firstTail := string([]rune(chunks[0])[overlapStart:])
	if !strings.HasPrefix(chunks[1], firstTail) {
		t.Fatalf("expected second chunk to start with %d-char overlap", defaultChunkOverlap)
	}
}

func TestImportPagesEmbedsEveryChunkAndStoresRecord(t *testing.T) {
	repo := &fakeRepository{}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	page := Page{
		Title:     "校园网 VPN 使用说明",
		Content:   strings.Repeat("VPN 登录后可以访问校内资源。", 90),
		Category:  "信息化服务",
		SourceURL: "https://nic.scut.edu.cn/vpn",
	}

	count, err := ImportPages(context.Background(), repo, embedder, []Page{page})
	if err != nil {
		t.Fatalf("import pages: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
	if len(repo.records) != 1 {
		t.Fatalf("expected one record, got %d", len(repo.records))
	}
	record := repo.records[0]
	if len(record.Chunks) != len(embedder.texts) {
		t.Fatalf("expected each chunk to be embedded, chunks=%d texts=%d", len(record.Chunks), len(embedder.texts))
	}
	for i, chunk := range record.Chunks {
		if embedder.texts[i] != chunk.Text {
			t.Fatalf("expected embedded text to match chunk %d", i)
		}
		if chunk.Embedding != "[0.10000000,0.20000000,0.30000000]" {
			t.Fatalf("unexpected embedding for chunk %d: %s", i, chunk.Embedding)
		}
	}
}

func TestImportPagesEmbedsChunksInSmallBatches(t *testing.T) {
	repo := &fakeRepository{}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	page := Page{
		Title:     "长文档",
		Content:   strings.Repeat("校园生活服务说明。", 220),
		SourceURL: "https://example.edu/long",
	}

	_, err := ImportPages(context.Background(), repo, embedder, []Page{page})
	if err != nil {
		t.Fatalf("import pages: %v", err)
	}
	if len(embedder.batchSizes) < 2 {
		t.Fatalf("expected multiple embed batches, got %#v", embedder.batchSizes)
	}
	for _, size := range embedder.batchSizes {
		if size > defaultEmbedBatchSize {
			t.Fatalf("expected batch size <= %d, got %d", defaultEmbedBatchSize, size)
		}
	}
}

func TestImportPagesProcessesRecordsInPageBatches(t *testing.T) {
	repo := &fakeRepository{}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	pages := []Page{
		{Title: "文档一", Content: "正文一", SourceURL: "https://example.edu/1"},
		{Title: "文档二", Content: "正文二", SourceURL: "https://example.edu/2"},
		{Title: "文档三", Content: "正文三", SourceURL: "https://example.edu/3"},
	}

	count, err := ImportPagesWithOptions(context.Background(), repo, embedder, pages, ImportOptions{
		PageBatchSize:  2,
		EmbedBatchSize: 2,
		ChunkBatchSize: 9,
	})
	if err != nil {
		t.Fatalf("import pages: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}
	if got, want := repo.batchSizes, []int{2, 1}; !equalInts(got, want) {
		t.Fatalf("expected repository batches %#v, got %#v", want, got)
	}
	if got, want := repo.chunkBatchSizes, []int{9, 9}; !equalInts(got, want) {
		t.Fatalf("expected chunk batch sizes %#v, got %#v", want, got)
	}
}

func TestImportPagesReportsBatchProgress(t *testing.T) {
	repo := &fakeRepository{}
	embedder := &fakeEmbedder{embedding: []float64{0.1, 0.2, 0.3}}
	var events []ProgressEvent
	pages := []Page{
		{Title: "文档一", Content: strings.Repeat("正文一", 120), SourceURL: "https://example.edu/1"},
		{Title: "文档二", Content: "正文二", SourceURL: "https://example.edu/2"},
	}

	_, err := ImportPagesWithOptions(context.Background(), repo, embedder, pages, ImportOptions{
		PageBatchSize:  2,
		EmbedBatchSize: 2,
		ChunkBatchSize: 9,
		OnProgress: func(event ProgressEvent) {
			events = append(events, event)
		},
	})
	if err != nil {
		t.Fatalf("import pages: %v", err)
	}

	wantStages := []ProgressStage{ProgressBatchStarted, ProgressBatchEmbedded, ProgressBatchImported}
	if len(events) != len(wantStages) {
		t.Fatalf("expected %d progress events, got %#v", len(wantStages), events)
	}
	for i, want := range wantStages {
		if events[i].Stage != want {
			t.Fatalf("expected event %d stage %q, got %#v", i, want, events[i])
		}
	}
	if events[0].TotalPages != 2 || events[0].BatchStart != 0 || events[0].BatchEnd != 1 {
		t.Fatalf("unexpected first event: %#v", events[0])
	}
	if events[1].ChunkCount == 0 {
		t.Fatalf("expected embedded event to include chunk count: %#v", events[1])
	}
	if events[2].ImportedCount != 2 {
		t.Fatalf("expected imported event count 2, got %#v", events[2])
	}
}

type fakeEmbedder struct {
	texts      []string
	batchSizes []int
	embedding  []float64
}

func (e *fakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	e.texts = append(e.texts, texts...)
	e.batchSizes = append(e.batchSizes, len(texts))
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embeddings[i] = e.embedding
	}
	return embeddings, nil
}

type fakeRepository struct {
	records         []qaimport.KnowledgeRecord
	batchSizes      []int
	chunkBatchSizes []int
}

func (r *fakeRepository) UpsertKnowledgeBatch(ctx context.Context, records []qaimport.KnowledgeRecord, options qaimport.BatchOptions) (int, error) {
	r.records = append(r.records, records...)
	r.batchSizes = append(r.batchSizes, len(records))
	r.chunkBatchSizes = append(r.chunkBatchSizes, options.ChunkBatchSize)
	return len(records), nil
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
