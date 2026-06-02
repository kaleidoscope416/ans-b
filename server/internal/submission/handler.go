package submission

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"ans-b/server/internal/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterStudentRoutes(group *gin.RouterGroup) {
	group.POST("", h.Create)
}

func (h *Handler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.POST("/:id/approve", h.Approve)
	group.POST("/:id/reject", h.Reject)
}

func (h *Handler) Create(c *gin.Context) {
	claims, ok := auth.CurrentUser(c)
	if !ok || claims.Role != auth.RoleStudent {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40001,
			"message": "missing current user",
			"data":    nil,
		})
		return
	}

	var request struct {
		Question string   `json:"question"`
		Answer   string   `json:"answer"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
		Source   string   `json:"source"`
		Remark   string   `json:"remark"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid request body",
			"data":    nil,
		})
		return
	}

	created, err := h.service.Create(c.Request.Context(), claims.UserID, CreateInput{
		Question: request.Question,
		Answer:   request.Answer,
		Category: request.Category,
		Tags:     request.Tags,
		Source:   request.Source,
		Remark:   request.Remark,
	})
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
		"data":    created,
	})
}

func (h *Handler) List(c *gin.Context) {
	claims, ok := auth.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40001,
			"message": "missing current user",
			"data":    nil,
		})
		return
	}

	var (
		submissions []Submission
		err         error
	)
	switch claims.Role {
	case auth.RoleStudent:
		submissions, err = h.service.ListForStudent(c.Request.Context(), claims.UserID)
	case auth.RoleAdmin:
		submissions, err = h.service.ListForAdmin(c.Request.Context(), c.Query("status"))
	default:
		c.JSON(http.StatusForbidden, gin.H{
			"code":    40003,
			"message": "permission denied",
			"data":    nil,
		})
		return
	}
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
		"data":    submissions,
	})
}

func (h *Handler) Approve(c *gin.Context) {
	claims, ok := auth.CurrentUser(c)
	if !ok || claims.Role != auth.RoleAdmin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40001,
			"message": "missing current user",
			"data":    nil,
		})
		return
	}

	submissionID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid submission id",
			"data":    nil,
		})
		return
	}

	var request struct {
		ReviewerNote string `json:"reviewer_note"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid request body",
			"data":    nil,
		})
		return
	}

	err = h.service.Approve(c.Request.Context(), submissionID, ReviewInput{
		ReviewerNote: request.ReviewerNote,
	})
	if err != nil {
		h.writeReviewError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":     submissionID,
			"status": StatusApproved,
		},
	})
}

func (h *Handler) Reject(c *gin.Context) {
	claims, ok := auth.CurrentUser(c)
	if !ok || claims.Role != auth.RoleAdmin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    40001,
			"message": "missing current user",
			"data":    nil,
		})
		return
	}

	submissionID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid submission id",
			"data":    nil,
		})
		return
	}

	var request struct {
		ReviewerNote string `json:"reviewer_note"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40000,
			"message": "invalid request body",
			"data":    nil,
		})
		return
	}

	err = h.service.Reject(c.Request.Context(), submissionID, ReviewInput{
		ReviewerNote: request.ReviewerNote,
	})
	if err != nil {
		h.writeReviewError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":     submissionID,
			"status": StatusRejected,
		},
	})
}

func (h *Handler) writeReviewError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	code := 40000
	message := err.Error()
	if errors.Is(err, sql.ErrNoRows) || strings.Contains(message, "not found") {
		status = http.StatusNotFound
		code = 40400
	}
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
