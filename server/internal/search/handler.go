package search

import (
	"ans-b/server/internal/httpresponse"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/candidates", h.Candidates)
}

func (h *Handler) Candidates(c *gin.Context) {
	// TODO: return ranked keyword and semantic search candidates.
	httpresponse.TODO(c, "search candidates")
}
