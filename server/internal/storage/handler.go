package storage

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
	group.POST("/imports", h.UploadImportFile)
}

func (h *Handler) UploadImportFile(c *gin.Context) {
	// TODO: validate file type and size, then persist import files.
	httpresponse.TODO(c, "import file storage")
}
