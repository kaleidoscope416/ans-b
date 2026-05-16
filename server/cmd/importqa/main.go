package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

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
	file := flag.String("file", "../data/qa_seed.json", "QA JSON file path")
	embedURL := flag.String("embed-url", defaultEmbedURL, "embedding service base URL")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
	count, err := qaimport.ImportFile(ctx, repo, embedder, *file)
	if err != nil {
		log.Fatalf("import qa: %v", err)
	}

	fmt.Printf("imported %d QA records from %s\n", count, *file)
}
