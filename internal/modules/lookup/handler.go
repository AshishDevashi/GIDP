package lookup

import (
	"net/http"

	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the lookup module's repository, service, and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	lookups := rg.Group("/lookups")
	lookups.GET("", m.all)
	lookups.GET("/repo-providers", m.repoProviders)
	lookups.GET("/languages", m.languages)
	lookups.GET("/repo-templates", m.repoTemplates)
}

func (m *Module) all(c *gin.Context) {
	items, err := m.service.All(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, items)
}

func (m *Module) repoProviders(c *gin.Context) {
	items, err := m.service.RepoProviders(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, items)
}

func (m *Module) languages(c *gin.Context) {
	items, err := m.service.Languages(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, items)
}

func (m *Module) repoTemplates(c *gin.Context) {
	items, err := m.service.RepoTemplates(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, items)
}
