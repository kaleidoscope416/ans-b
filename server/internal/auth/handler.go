package auth

import (
	"errors"
	"net/http"
	"strings"

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
	group.POST("/logout", h.Logout)
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

func (h *Handler) Logout(c *gin.Context) {
	token, err := BearerToken(c.GetHeader("Authorization"))
	if err != nil {
		unauthorized(c, "missing authorization token")
		return
	}
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		if isTokenError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40001, "message": err.Error(), "data": nil})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 50000, "message": "session store unavailable", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": nil})
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

func isTokenError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "token") || strings.Contains(message, "unsupported")
}
