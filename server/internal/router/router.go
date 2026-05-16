package router

import (
	"database/sql"
	"net/http"
	"strings"

	"ans-b/server/internal/analytics"
	"ans-b/server/internal/auth"
	"ans-b/server/internal/knowledge"
	"ans-b/server/internal/model"
	"ans-b/server/internal/qa"
	"ans-b/server/internal/search"
	"ans-b/server/internal/storage"
	"ans-b/server/internal/submission"
	"ans-b/server/internal/user"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	RegisterRoutesWithDBAndEmbedder(engine, nil, nil)
}

func RegisterRoutesWithDB(engine *gin.Engine, db *sql.DB) {
	RegisterRoutesWithDBAndEmbedder(engine, db, nil)
}

func RegisterRoutesWithDBAndEmbedder(engine *gin.Engine, db *sql.DB, embedder qa.Embedder) {
	engine.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "null" ||
			strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			strings.HasPrefix(origin, "http://100.") {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	api := engine.Group("/api/v1")

	auth.NewHandler(auth.NewService(auth.NewRepository())).RegisterRoutes(api.Group("/auth"))
	user.NewHandler(user.NewService(user.NewRepository())).RegisterRoutes(api.Group("/users"))
	knowledge.NewHandler(knowledge.NewService(knowledge.NewRepository(db), embedder)).RegisterRoutes(api.Group("/knowledge"))
	qa.NewHandler(qa.NewService(qa.NewRepository(db), embedder)).RegisterRoutes(api.Group("/qa"))
	search.NewHandler(search.NewService(search.NewRepository())).RegisterRoutes(api.Group("/search"))
	submission.NewHandler(submission.NewService(submission.NewRepository())).RegisterRoutes(api.Group("/submissions"))
	analytics.NewHandler(analytics.NewService(analytics.NewRepository())).RegisterRoutes(api.Group("/analytics"))
	model.NewHandler(model.NewService(model.NewRepository())).RegisterRoutes(api.Group("/model"))
	storage.NewHandler(storage.NewService(storage.NewRepository())).RegisterRoutes(api.Group("/storage"))
}
