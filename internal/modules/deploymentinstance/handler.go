package deploymentinstance

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

type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool, cfg Config, log *slog.Logger) *Module {
	repo := NewRepository(pool)

	var prov *provisioner
	runner, err := terraform.NewRunner(cfg.TerraformBinPath, cfg.WorkDir, nil)
	if err != nil {
		log.Warn("deployment instance provisioning disabled", "error", err)
	} else if prov, err = newProvisioner(runner, cfg); err != nil {
		log.Warn("deployment instance provisioning disabled", "error", err)
		prov = nil
	}

	return &Module{service: NewService(repo, prov, cfg, log)}
}

func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	instances := rg.Group("/deployment-instances")
	instances.POST("", m.create)
	instances.GET("", m.get)
	instances.DELETE("", m.delete)
}

func (m *Module) create(c *gin.Context) {
	instance, err := m.service.Create(c.Request.Context(), c.GetString(auth.ContextUserIDKey))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusAccepted, instance)
}

func (m *Module) get(c *gin.Context) {
	instance, err := m.service.Get(c.Request.Context())
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, instance)
}

func (m *Module) delete(c *gin.Context) {
	if err := m.service.Delete(c.Request.Context()); err != nil {
		m.respondServiceError(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrBusy), errors.Is(err, ErrAlreadyExists), errors.Is(err, ErrActiveDeployments):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotConfigured):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "deployment instance operation failed")
	}
}
