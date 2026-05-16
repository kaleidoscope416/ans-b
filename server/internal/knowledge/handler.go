package knowledge

import (
	"net/http"

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
	group.GET("", h.List)
	group.POST("", h.Create)
	group.GET("/:id", h.Get)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
	group.POST("/import", h.Import)
}

func (h *Handler) List(c *gin.Context) {
	// TODO: list knowledge base entries with filters and pagination.
	httpresponse.TODO(c, "knowledge list")
}

func (h *Handler) Create(c *gin.Context) {
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

	err := h.service.Create(c.Request.Context(), CreateInput{
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
		"data": gin.H{
			"question": request.Question,
			"category": request.Category,
		},
	})
}

func (h *Handler) Get(c *gin.Context) {
	// TODO: return a single knowledge base entry by ID.
	httpresponse.TODO(c, "knowledge detail")
}

func (h *Handler) Update(c *gin.Context) {
	// TODO: validate and update a knowledge base entry.
	httpresponse.TODO(c, "knowledge update")
}

func (h *Handler) Delete(c *gin.Context) {
	// TODO: delete or archive a knowledge base entry.
	httpresponse.TODO(c, "knowledge delete")
}

func (h *Handler) Import(c *gin.Context) {
	// TODO: parse and import FAQ files after admin authorization.
	httpresponse.TODO(c, "knowledge import")
}
