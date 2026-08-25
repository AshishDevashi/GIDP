package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Module exposes a simple liveness/readiness endpoint.
type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", m.check)
}

func (m *Module) check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
