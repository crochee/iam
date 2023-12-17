package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}
