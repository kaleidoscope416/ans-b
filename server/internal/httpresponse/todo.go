package httpresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO returns a consistent placeholder response for scaffolded endpoints.
func TODO(c *gin.Context, module string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    "TODO",
		"message": "TODO: implement " + module + " server logic",
	})
}
