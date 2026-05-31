package docimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"ans-b/server/internal/embedding"
	"ans-b/server/internal/qaimport"
)

const (
	defaultChunkSize      = 300
	defaultChunkOverlap   = 60
	defaultEmbedBatchSize = 2
	defaultPageBatchSize  = 50
	defaultChunkBatchSize = 500
)

type Page struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Category   string   `json:"category"`
	SourceName string   `json:"source_name"`
	SourceURL  string   `json:"source_url"`
	Tags       []string `json:"tags"`
}

type Repository interface {
	UpsertKnowledgeBatch(ctx context.Context, records []qaimport.KnowledgeRecord, options qaimport.BatchOptions) (int, error)
}

type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

type ImportOptions struct {
	PageBatchSize  int
	EmbedBatchSize int
	ChunkBatchSize int
	OnProgress     func(ProgressEvent)
}

type ProgressStage string

const (
	ProgressBatchStarted  ProgressStage = "batch_started"
	ProgressBatchEmbedded ProgressStage = "batch_embedded"
	ProgressBatchImported ProgressStage = "batch_imported"
)

type ProgressEvent struct {
	Stage         ProgressStage
	BatchStart    int
	BatchEnd      int
	TotalPages    int
	PageCount     int
	ChunkCount    int
	ImportedCount int
}

func LoadFile(path string) ([]Page, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pages []Page
	if err := json.Unmarshal(data, &pages); err != nil {
		return nil, err
	}
	return pages, nil
}

func ImportFile(ctx context.Context, repo Repository, embedder Embedder, path string) (int, error) {
	return ImportFileWithOptions(ctx, repo, embedder, path, ImportOptions{})
}

func ImportFileWithOptions(ctx context.Context, repo Repository, embedder Embedder, path string, options ImportOptions) (int, error) {
	pages, err := LoadFile(path)
	if err != nil {
		return 0, fmt.Errorf("load document file: %w", err)
	}
	return ImportPagesWithOptions(ctx, repo, embedder, pages, options)
}

func ImportPages(ctx context.Context, repo Repository, embedder Embedder, pages []Page) (int, error) {
	return ImportPagesWithOptions(ctx, repo, embedder, pages, ImportOptions{})
}

func ImportPagesWithOptions(ctx context.Context, repo Repository, embedder Embedder, pages []Page, options ImportOptions) (int, error) {
	if repo == nil {
		return 0, errors.New("repository is required")
	}
	if embedder == nil {
		return 0, errors.New("embedder is required")
	}
	options = normalizeImportOptions(options)

	count := 0
	for start := 0; start < len(pages); start += options.PageBatchSize {
		end := start + options.PageBatchSize
		if end > len(pages) {
			end = len(pages)
		}
		imported, err := importPageBatch(ctx, repo, embedder, pages[start:end], batchContext{
			start: start,
			end:   end,
			total: len(pages),
		}, options)
		if err != nil {
			return count, fmt.Errorf("import batch %d-%d: %w", start, end-1, err)
		}
		count += imported
	}
	return count, nil
}

type batchContext struct {
	start int
	end   int
	total int
}

func importPageBatch(ctx context.Context, repo Repository, embedder Embedder, pages []Page, batch batchContext, options ImportOptions) (int, error) {
	records := make([]qaimport.KnowledgeRecord, 0, len(pages))
	chunkTexts := make([]string, 0)
	for _, page := range pages {
		record, err := BuildRecord(page)
		if err != nil {
			return 0, err
		}
		records = append(records, record)
		for _, chunk := range record.Chunks {
			chunkTexts = append(chunkTexts, chunk.Text)
		}
	}

	reportProgress(options, ProgressEvent{
		Stage:      ProgressBatchStarted,
		BatchStart: batch.start,
		BatchEnd:   batch.end - 1,
		TotalPages: batch.total,
		PageCount:  len(pages),
		ChunkCount: len(chunkTexts),
	})

	embeddings, err := embedInBatches(ctx, embedder, chunkTexts, options.EmbedBatchSize)
	if err != nil {
		return 0, fmt.Errorf("embed document chunks: %w", err)
	}
	if len(embeddings) != len(chunkTexts) {
		return 0, fmt.Errorf("embedding count mismatch: got %d, want %d", len(embeddings), len(chunkTexts))
	}
	reportProgress(options, ProgressEvent{
		Stage:      ProgressBatchEmbedded,
		BatchStart: batch.start,
		BatchEnd:   batch.end - 1,
		TotalPages: batch.total,
		PageCount:  len(pages),
		ChunkCount: len(chunkTexts),
	})

	embeddingIndex := 0
	for i := range records {
		for j := range records[i].Chunks {
			records[i].Chunks[j].Embedding = embedding.VectorLiteral(embeddings[embeddingIndex])
			embeddingIndex++
		}
	}
	imported, err := repo.UpsertKnowledgeBatch(ctx, records, qaimport.BatchOptions{
		ChunkBatchSize: options.ChunkBatchSize,
	})
	if err != nil {
		return 0, err
	}
	reportProgress(options, ProgressEvent{
		Stage:         ProgressBatchImported,
		BatchStart:    batch.start,
		BatchEnd:      batch.end - 1,
		TotalPages:    batch.total,
		PageCount:     len(pages),
		ChunkCount:    len(chunkTexts),
		ImportedCount: imported,
	})
	return imported, nil
}

func reportProgress(options ImportOptions, event ProgressEvent) {
	if options.OnProgress != nil {
		options.OnProgress(event)
	}
}

func normalizeImportOptions(options ImportOptions) ImportOptions {
	if options.PageBatchSize <= 0 {
		options.PageBatchSize = defaultPageBatchSize
	}
	if options.EmbedBatchSize <= 0 {
		options.EmbedBatchSize = defaultEmbedBatchSize
	}
	if options.ChunkBatchSize <= 0 {
		options.ChunkBatchSize = defaultChunkBatchSize
	}
	return options
}

func embedInBatches(ctx context.Context, embedder Embedder, texts []string, batchSize int) ([][]float64, error) {
	if batchSize <= 0 {
		batchSize = defaultEmbedBatchSize
	}
	embeddings := make([][]float64, 0, len(texts))
	for start := 0; start < len(texts); start += batchSize {
		end := start + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batchEmbeddings, err := embedder.Embed(ctx, texts[start:end])
		if err != nil {
			return nil, err
		}
		if len(batchEmbeddings) != end-start {
			return nil, fmt.Errorf("embedding count mismatch: got %d, want %d", len(batchEmbeddings), end-start)
		}
		embeddings = append(embeddings, batchEmbeddings...)
	}
	return embeddings, nil
}

func BuildRecord(page Page) (qaimport.KnowledgeRecord, error) {
	if err := validatePage(page); err != nil {
		return qaimport.KnowledgeRecord{}, err
	}
	title := strings.TrimSpace(page.Title)
	content := normalizeContent(page.Content)
	sourceURL := strings.TrimSpace(page.SourceURL)

	chunks := SplitContent(content)
	recordChunks := make([]qaimport.KnowledgeChunk, 0, len(chunks))
	for _, chunk := range chunks {
		recordChunks = append(recordChunks, qaimport.KnowledgeChunk{
			Text:      chunk,
			SourceURL: sourceURL,
		})
	}

	return qaimport.KnowledgeRecord{
		Title:      title,
		Question:   title,
		Answer:     content,
		Category:   strings.TrimSpace(page.Category),
		Tags:       cleanTags(page.Tags),
		Source:     strings.TrimSpace(page.SourceName),
		SourceType: "official_page",
		Chunks:     recordChunks,
	}, nil
}

func SplitContent(content string) []string {
	return splitContent(content, defaultChunkSize, defaultChunkOverlap)
}

func splitContent(content string, chunkSize int, overlap int) []string {
	content = normalizeContent(content)
	if content == "" {
		return nil
	}
	runes := []rune(content)
	if chunkSize <= 0 || len(runes) <= chunkSize {
		return []string{content}
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 4
	}

	chunks := make([]string, 0, len(runes)/chunkSize+1)
	for start := 0; start < len(runes); {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
		start = end - overlap
	}
	return chunks
}

func validatePage(page Page) error {
	if strings.TrimSpace(page.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(page.Content) == "" {
		return errors.New("content is required")
	}
	if strings.TrimSpace(page.SourceURL) == "" {
		return errors.New("source_url is required")
	}
	return nil
}

func normalizeContent(value string) string {
	lines := strings.Fields(strings.TrimSpace(value))
	return strings.Join(lines, " ")
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
