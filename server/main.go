package main

import (
	"database/sql"
	"log"
	"os"

	"ans-b/server/internal/router"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
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

	engine := gin.Default()
	router.RegisterRoutesWithDB(engine, db)

	if err := engine.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
