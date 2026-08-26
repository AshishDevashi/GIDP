package service

import (
	"errors"
	"net/http"

	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the service module's repository, service, and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	services := rg.Group("/services")
	services.POST("", m.create)
	services.GET("", m.list)
	services.GET("/:id", m.getBySlug)
	services.POST("/:id/environments", m.addEnvironment)
	services.GET("/:id/environments", m.listEnvironments)
	services.POST("/:id/dependencies", m.addDependency)
	services.GET("/:id/dependencies", m.listDependencies)
	services.GET("/:id/dependents", m.listDependents)
	services.DELETE("/:id/dependencies/:dependsOnId", m.removeDependency)
	services.POST("/:id/tags", m.addTag)
	services.GET("/:id/tags", m.listTags)
	services.DELETE("/:id/tags/:tag", m.removeTag)
}

func (m *Module) create(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	svc, err := m.service.Create(c.Request.Context(), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, svc)
}

func (m *Module) list(c *gin.Context) {
	services, err := m.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, services)
}

func (m *Module) getBySlug(c *gin.Context) {
	svc, err := m.service.GetBySlug(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, svc)
}

func (m *Module) addEnvironment(c *gin.Context) {
	var req AddEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	env, err := m.service.AddEnvironment(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, env)
}

func (m *Module) listEnvironments(c *gin.Context) {
	envs, err := m.service.ListEnvironments(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, envs)
}

func (m *Module) addDependency(c *gin.Context) {
	var req AddDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	dep, err := m.service.AddDependency(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, dep)
}

func (m *Module) listDependencies(c *gin.Context) {
	deps, err := m.service.ListDependencies(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, deps)
}

// listDependents shows what would be affected if this service were deleted.
func (m *Module) listDependents(c *gin.Context) {
	deps, err := m.service.ListDependents(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, deps)
}

func (m *Module) removeDependency(c *gin.Context) {
	if err := m.service.RemoveDependency(c.Request.Context(), c.Param("id"), c.Param("dependsOnId")); err != nil {
		m.respondServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (m *Module) addTag(c *gin.Context) {
	var req AddTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := m.service.AddTag(c.Request.Context(), c.Param("id"), req); err != nil {
		m.respondServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (m *Module) listTags(c *gin.Context) {
	tags, err := m.service.ListTags(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, tags)
}

func (m *Module) removeTag(c *gin.Context) {
	if err := m.service.RemoveTag(c.Request.Context(), c.Param("id"), c.Param("tag")); err != nil {
		m.respondServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSlugTaken), errors.Is(err, ErrEnvironmentTaken):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
