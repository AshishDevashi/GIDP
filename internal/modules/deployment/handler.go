package deployment

import (
	"errors"
	"net/http"

	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the deployment module's repository, service, and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	deployments := rg.Group("/deployments")
	deployments.POST("", m.create)
	deployments.GET("/:id", m.getByID)
	deployments.PATCH("/:id/status", m.updateStatus)

	serviceDeployments := rg.Group("/services/:serviceId/deployments")
	serviceDeployments.GET("", m.listByService)
	serviceDeployments.GET("/current", m.current)
}

func (m *Module) create(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	dep, err := m.service.Create(c.Request.Context(), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, dep)
}

func (m *Module) getByID(c *gin.Context) {
	dep, err := m.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, dep)
}

func (m *Module) updateStatus(c *gin.Context) {
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	dep, err := m.service.UpdateStatus(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, dep)
}

func (m *Module) listByService(c *gin.Context) {
	deps, err := m.service.ListByService(c.Request.Context(), c.Param("serviceId"), c.Query("environment"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, deps)
}

// current returns what's actually running right now for a service+environment.
func (m *Module) current(c *gin.Context) {
	environment := c.Query("environment")
	if environment == "" {
		response.Error(c, http.StatusBadRequest, "environment query parameter is required")
		return
	}

	dep, err := m.service.Current(c.Request.Context(), c.Param("serviceId"), environment)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, dep)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
