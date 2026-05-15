package main

import (
	"log"

	"ans-b/server/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	router.RegisterRoutes(engine)

	if err := engine.Run(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
