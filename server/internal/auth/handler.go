package auth

import (
	"errors"
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
	group.POST("/student/login", h.StudentLogin)
	group.POST("/admin/login", h.AdminLogin)
}

func (h *Handler) StudentLogin(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "invalid request body", "data": nil})
		return
	}
	result, err := h.service.LoginStudent(c.Request.Context(), LoginInput{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		status, code := loginErrorStatus(err)
		c.JSON(status, gin.H{"code": code, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}

func (h *Handler) AdminLogin(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "invalid request body", "data": nil})
		return
	}
	result, err := h.service.LoginAdmin(c.Request.Context(), LoginInput{
		Username: request.Username,
		Password: request.Password,
	})
	if err != nil {
		status, code := loginErrorStatus(err)
		c.JSON(status, gin.H{"code": code, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": result})
}

func loginErrorStatus(err error) (int, int) {
	if errors.Is(err, ErrInvalidLoginInput) {
		return http.StatusBadRequest, 40000
	}
	if errors.Is(err, ErrInvalidCredentials) {
		return http.StatusUnauthorized, 40001
	}
	return http.StatusInternalServerError, 50000
}
