package qaimport

import (
	"strings"
	"testing"
)

func TestBuildChunkInsertQueryUsesMultiValueInsert(t *testing.T) {
	query, args := buildChunkInsertQuery([]chunkInsertRow{
		{
			itemID: 10,
			chunk: KnowledgeChunk{
				Text:      "第一段",
				Embedding: "[0.10000000,0.20000000]",
				SourceURL: "https://example.edu/1",
			},
		},
		{
			itemID: 11,
			chunk: KnowledgeChunk{
				Text:      "第二段",
				Embedding: "[0.30000000,0.40000000]",
				SourceURL: "https://example.edu/2",
			},
		},
	})

	if strings.Count(query, "VALUES") != 1 {
		t.Fatalf("expected one VALUES clause, got query:\n%s", query)
	}
	if !strings.Contains(query, "($1, $2, $3::vector, $4),($5, $6, $7::vector, $8)") {
		t.Fatalf("expected multi-row vector insert placeholders, got query:\n%s", query)
	}
	if len(args) != 8 {
		t.Fatalf("expected 8 args, got %d", len(args))
	}
	if args[0] != int64(10) || args[1] != "第一段" || args[4] != int64(11) || args[5] != "第二段" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestNormalizeBatchOptionsUsesDefaultChunkBatchSize(t *testing.T) {
	options := normalizeBatchOptions(BatchOptions{})

	if options.ChunkBatchSize != defaultChunkInsertBatchSize {
		t.Fatalf("expected default chunk batch size %d, got %d", defaultChunkInsertBatchSize, options.ChunkBatchSize)
	}
}
