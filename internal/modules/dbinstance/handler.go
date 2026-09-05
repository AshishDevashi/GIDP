package dbinstance

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/AshishDevashi/GIDP/internal/modules/auth"
	"github.com/AshishDevashi/GIDP/internal/platform/terraform"
	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the db instance module's service and HTTP routes together.
type Module struct {
	service *Service
}

// NewModule builds the module. When Terraform is unavailable the module still
// serves read routes and reports provisioning as unconfigured.
func NewModule(pool *pgxpool.Pool, cfg Config, log *slog.Logger) *Module {
	repo := NewRepository(pool)

	var prov *provisioner
	runner, err := terraform.NewRunner(cfg.TerraformBinPath, cfg.WorkDir, nil)
	if err != nil {
		log.Warn("db instance provisioning disabled", "error", err)
	} else if prov, err = newProvisioner(runner, cfg); err != nil {
		log.Warn("db instance provisioning disabled", "error", err)
		prov = nil
	}

	return &Module{service: NewService(repo, prov, cfg, log)}
}

// Service returns the underlying service instance.
func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	instances := rg.Group("/db-instances")
	instances.POST("", m.create)
	instances.GET("", m.list)
	instances.GET("/:id", m.getByID)
	instances.DELETE("/:id", m.delete)
}

func (m *Module) create(c *gin.Context) {
	instance, err := m.service.Create(c.Request.Context(), c.GetString(auth.ContextUserIDKey))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusAccepted, instance)
}

func (m *Module) list(c *gin.Context) {
	instances, err := m.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list db instances")
		return
	}
	response.JSON(c, http.StatusOK, instances)
}

func (m *Module) getByID(c *gin.Context) {
	instance, err := m.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, instance)
}

func (m *Module) delete(c *gin.Context) {
	if err := m.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		m.respondServiceError(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrBusy), errors.Is(err, ErrAlreadyExists):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotConfigured):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "db instance operation failed")
	}
}
