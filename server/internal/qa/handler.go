package qa

import (
	"net/http"

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
	var request struct {
		Question string `json:"question"`
		Limit    int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid request body",
			"data":    nil,
		})
		return
	}

	answer, err := h.service.Ask(c.Request.Context(), request.Question, request.Limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    answer,
	})
}
