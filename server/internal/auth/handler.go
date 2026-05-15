package auth

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
	group.POST("/student/login", h.StudentLogin)
	group.POST("/admin/login", h.AdminLogin)
}

func (h *Handler) StudentLogin(c *gin.Context) {
	// TODO: validate student credentials and return an auth token.
	httpresponse.TODO(c, "student login")
}

func (h *Handler) AdminLogin(c *gin.Context) {
	// TODO: validate administrator credentials and return an auth token.
	httpresponse.TODO(c, "admin login")
}
