package submission

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
	group.POST("", h.Create)
	group.GET("", h.List)
	group.POST("/:id/approve", h.Approve)
	group.POST("/:id/reject", h.Reject)
}

func (h *Handler) Create(c *gin.Context) {
	// TODO: authenticate the student and save a pending submission.
	httpresponse.TODO(c, "submission create")
}

func (h *Handler) List(c *gin.Context) {
	// TODO: list submissions for students or administrators based on authorization.
	httpresponse.TODO(c, "submission list")
}

func (h *Handler) Approve(c *gin.Context) {
	// TODO: approve a submission and publish it into the knowledge base.
	httpresponse.TODO(c, "submission approve")
}

func (h *Handler) Reject(c *gin.Context) {
	// TODO: reject a submission with an audit remark.
	httpresponse.TODO(c, "submission reject")
}
