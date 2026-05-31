package model

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
	group.POST("/embeddings", h.CreateEmbedding)
	group.POST("/chat", h.Chat)
}

func (h *Handler) CreateEmbedding(c *gin.Context) {
	// TODO: call the configured embedding model provider.
	httpresponse.TODO(c, "embedding model")
}

func (h *Handler) Chat(c *gin.Context) {
	// TODO: call the optional large language model provider.
	httpresponse.TODO(c, "large language model")
}
