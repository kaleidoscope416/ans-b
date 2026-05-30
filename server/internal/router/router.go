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
	RegisterRoutesWithDBEmbedderAndSessionStore(engine, nil, nil, auth.NewMemorySessionStore())
}

func RegisterRoutesWithDB(engine *gin.Engine, db *sql.DB) {
	RegisterRoutesWithDBEmbedderAndSessionStore(engine, db, nil, auth.NewMemorySessionStore())
}

func RegisterRoutesWithDBAndEmbedder(engine *gin.Engine, db *sql.DB, embedder qa.Embedder, generators ...qa.AnswerGenerator) {
	RegisterRoutesWithDBEmbedderAndSessionStore(engine, db, embedder, auth.NewMemorySessionStore(), generators...)
}

func RegisterRoutesWithDBEmbedderAndSessionStore(engine *gin.Engine, db *sql.DB, embedder qa.Embedder, sessionStore auth.SessionStore, generators ...qa.AnswerGenerator) {
	var generator qa.AnswerGenerator
	if len(generators) > 0 {
		generator = generators[0]
	}

	engine.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "null" ||
			strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			strings.HasPrefix(origin, "http://100.") {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

	tokenManager := auth.NewTokenManagerFromEnv()
	auth.NewHandler(auth.NewService(auth.NewRepository(db), tokenManager, sessionStore)).RegisterRoutes(api.Group("/auth"))

	userHandler := user.NewHandler(user.NewService(user.NewRepository(db)))
	userHandler.RegisterRoutes(api.Group("/users"))
	userHandler.RegisterProtectedRoutes(api.Group("/users", auth.Middleware(tokenManager, sessionStore, auth.RoleStudent)))

	analyticsService := analytics.NewService(analytics.NewRepository(db))

	knowledge.NewHandler(knowledge.NewService(knowledge.NewRepository(db), embedder)).RegisterRoutes(api.Group("/knowledge"))
	qaService := qa.NewService(qa.NewRepository(db), embedder, generator)
	qaService.SetAccessRecorder(analyticsService)
	qa.NewHandler(qaService).RegisterRoutes(api.Group("/qa", auth.Middleware(tokenManager, sessionStore, auth.RoleStudent)))
	search.NewHandler(search.NewService(search.NewRepository())).RegisterRoutes(api.Group("/search"))
	submission.NewHandler(submission.NewService(submission.NewRepository())).RegisterRoutes(api.Group("/submissions"))
	analytics.NewHandler(analyticsService).RegisterRoutes(api.Group("/analytics"))
	model.NewHandler(model.NewService(model.NewRepository())).RegisterRoutes(api.Group("/model"))
	storage.NewHandler(storage.NewService(storage.NewRepository())).RegisterRoutes(api.Group("/storage"))
}
