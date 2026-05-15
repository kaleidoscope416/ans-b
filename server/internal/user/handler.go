package user

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
	group.POST("/register", h.Register)
	group.GET("/me", h.Profile)
}

func (h *Handler) Register(c *gin.Context) {
	// TODO: validate registration input and create a student account.
	httpresponse.TODO(c, "student registration")
}

func (h *Handler) Profile(c *gin.Context) {
	// TODO: authenticate the student and return profile information.
	httpresponse.TODO(c, "student profile")
}
