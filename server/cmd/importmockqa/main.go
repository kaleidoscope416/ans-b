package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ans-b/server/internal/mockqa"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	defaultDBURL := os.Getenv("DATABASE_URL")
	if defaultDBURL == "" {
		defaultDBURL = "postgres://campus:campus123@localhost:5432/campus_qa?sslmode=disable"
	}

	dbURL := flag.String("db", defaultDBURL, "PostgreSQL connection URL")
	file := flag.String("file", "../data/mock_qa.json", "mock QA JSON file path")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", *dbURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	repo := mockqa.NewPostgresRepository(db)
	count, err := mockqa.ImportFile(ctx, repo, *file)
	if err != nil {
		log.Fatalf("import mock qa: %v", err)
	}

	fmt.Printf("imported %d mock QA records from %s\n", count, *file)
}
