package analytics

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
	group.GET("/hot-questions", h.HotQuestions)
}

func (h *Handler) HotQuestions(c *gin.Context) {
	// TODO: aggregate and return the most frequent questions.
	httpresponse.TODO(c, "hot question analytics")
}
