package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"ans-b/server/internal/auth"
	"ans-b/server/internal/config"
	"ans-b/server/internal/embedding"
	"ans-b/server/internal/llm"
	"ans-b/server/internal/router"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := config.LoadDotEnvFiles(".env", "../.env"); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://campus:campus123@localhost:5432/campus_qa?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	embedBaseURL := os.Getenv("EMBED_BASE_URL")
	if embedBaseURL == "" {
		embedBaseURL = "http://127.0.0.1:18080"
	}
	embedder := embedding.NewHTTPClient(embedBaseURL)
	tokenManager := auth.NewTokenManagerFromEnv()
	authService := auth.NewService(auth.NewRepository(db), tokenManager)
	if err := authService.InitAuthSystem(context.Background()); err != nil {
		log.Fatalf("failed to initialize auth system: %v", err)
	}
	answerGenerator := llm.NewOpenAICompatibleFromEnv()

	engine := gin.Default()
	router.RegisterRoutesWithDBAndEmbedder(engine, db, embedder, answerGenerator)

	if err := engine.Run(":23456"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
