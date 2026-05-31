package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ans-b/server/internal/docimport"
	"ans-b/server/internal/embedding"
	"ans-b/server/internal/qaimport"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	defaultDBURL := os.Getenv("DATABASE_URL")
	if defaultDBURL == "" {
		defaultDBURL = "postgres://campus:campus123@localhost:5432/campus_qa?sslmode=disable"
	}

	defaultEmbedURL := os.Getenv("EMBED_BASE_URL")
	if defaultEmbedURL == "" {
		defaultEmbedURL = "http://127.0.0.1:18080"
	}

	dbURL := flag.String("db", defaultDBURL, "PostgreSQL connection URL")
	file := flag.String("file", "../data/scut_pages.json", "document page JSON file path")
	embedURL := flag.String("embed-url", defaultEmbedURL, "embedding service base URL")
	pageBatchSize := flag.Int("page-batch-size", 50, "document records per database transaction")
	embedBatchSize := flag.Int("embed-batch-size", 2, "texts per embedding request")
	chunkBatchSize := flag.Int("chunk-batch-size", 500, "chunks per database insert statement")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	db, err := sql.Open("pgx", *dbURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	repo := qaimport.NewPostgresRepository(db)
	embedder := embedding.NewHTTPClient(*embedURL)
	log.Printf("starting document import file=%s page_batch_size=%d embed_batch_size=%d chunk_batch_size=%d", *file, *pageBatchSize, *embedBatchSize, *chunkBatchSize)
	count, err := docimport.ImportFileWithOptions(ctx, repo, embedder, *file, docimport.ImportOptions{
		PageBatchSize:  *pageBatchSize,
		EmbedBatchSize: *embedBatchSize,
		ChunkBatchSize: *chunkBatchSize,
		OnProgress: func(event docimport.ProgressEvent) {
			switch event.Stage {
			case docimport.ProgressBatchStarted:
				log.Printf("batch %d-%d/%d: built %d pages into %d chunks, embedding...", event.BatchStart+1, event.BatchEnd+1, event.TotalPages, event.PageCount, event.ChunkCount)
			case docimport.ProgressBatchEmbedded:
				log.Printf("batch %d-%d/%d: embedded %d chunks, writing database...", event.BatchStart+1, event.BatchEnd+1, event.TotalPages, event.ChunkCount)
			case docimport.ProgressBatchImported:
				log.Printf("batch %d-%d/%d: imported %d pages", event.BatchStart+1, event.BatchEnd+1, event.TotalPages, event.ImportedCount)
			}
		},
	})
	if err != nil {
		log.Fatalf("import documents: %v", err)
	}

	fmt.Printf("imported %d document records from %s\n", count, *file)
}
