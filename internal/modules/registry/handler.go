package registry

import (
	"errors"
	"net/http"

	"github.com/AshishDevashi/GIDP/internal/modules/auth"
	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the registry module's service and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool, dockerHubUsername, dockerHubToken, defaultNamespace string) *Module {
	repo := NewRepository(pool)
	client := newDockerHubClient(dockerHubUsername, dockerHubToken)
	return &Module{service: NewService(repo, client, defaultNamespace)}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	registries := rg.Group("/registries")
	registries.POST("", m.create)
	registries.GET("", m.list)
	registries.GET("/:id", m.getByID)
	registries.PUT("/:id", m.update)
	registries.DELETE("/:id", m.delete)
}

func (m *Module) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	registry, err := m.service.Create(c.Request.Context(), req, c.GetString(auth.ContextUserIDKey))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, registry)
}

func (m *Module) list(c *gin.Context) {
	registries, err := m.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list registries")
		return
	}
	response.JSON(c, http.StatusOK, registries)
}

func (m *Module) getByID(c *gin.Context) {
	registry, err := m.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, registry)
}

func (m *Module) update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	registry, err := m.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, registry)
}

func (m *Module) delete(c *gin.Context) {
	if err := m.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		m.respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNamespaceRequired), errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrRegistryTaken):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrRegistryInvalid):
		response.Error(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrUnauthorized):
		response.Error(c, http.StatusBadGateway, err.Error())
	case errors.Is(err, ErrNotConfigured), errors.Is(err, ErrDockerHubUnavailabe):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "registry operation failed")
	}
}
