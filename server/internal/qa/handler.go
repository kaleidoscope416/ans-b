package qa

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
	group.POST("/ask", h.Ask)
}

func (h *Handler) Ask(c *gin.Context) {
	// TODO: orchestrate retrieval, confidence handling, logging, and answer response.
	httpresponse.TODO(c, "question answering")
}
