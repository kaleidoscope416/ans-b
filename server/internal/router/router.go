package router

import (
	"net/http"

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
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	api := engine.Group("/api/v1")

	auth.NewHandler(auth.NewService(auth.NewRepository())).RegisterRoutes(api.Group("/auth"))
	user.NewHandler(user.NewService(user.NewRepository())).RegisterRoutes(api.Group("/users"))
	knowledge.NewHandler(knowledge.NewService(knowledge.NewRepository())).RegisterRoutes(api.Group("/knowledge"))
	qa.NewHandler(qa.NewService(qa.NewRepository())).RegisterRoutes(api.Group("/qa"))
	search.NewHandler(search.NewService(search.NewRepository())).RegisterRoutes(api.Group("/search"))
	submission.NewHandler(submission.NewService(submission.NewRepository())).RegisterRoutes(api.Group("/submissions"))
	analytics.NewHandler(analytics.NewService(analytics.NewRepository())).RegisterRoutes(api.Group("/analytics"))
	model.NewHandler(model.NewService(model.NewRepository())).RegisterRoutes(api.Group("/model"))
	storage.NewHandler(storage.NewService(storage.NewRepository())).RegisterRoutes(api.Group("/storage"))
}
