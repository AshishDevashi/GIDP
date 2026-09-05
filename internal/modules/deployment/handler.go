package deployment

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/AshishDevashi/GIDP/internal/modules/auth"
	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool, log *slog.Logger) *Module {
	repo := NewRepository(pool)
	return &Module{service: NewService(repo, log)}
}

func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	deployments := rg.Group("/deployments")
	deployments.POST("", m.create)
	deployments.GET("", m.list)
	deployments.GET("/:id", m.getByID)
	deployments.DELETE("/:id", m.delete)
}

func (m *Module) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	created, err := m.service.Create(c.Request.Context(), req, c.GetString(auth.ContextUserIDKey))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (m *Module) list(c *gin.Context) {
	deployments, err := m.service.List(c.Request.Context())
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, deployments)
}

func (m *Module) getByID(c *gin.Context) {
	deployment, err := m.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, deployment)
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
	case errors.Is(err, ErrInvalidID), errors.Is(err, ErrNoActiveDeploymentInstance), errors.Is(err, ErrInvalidManifestInput), errors.Is(err, ErrKubernetesApply):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrRegistryNotFound), errors.Is(err, ErrRepoNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrCapacityReached), errors.Is(err, ErrNameTaken):
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "deployment operation failed")
	}
}
