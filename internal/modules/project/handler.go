package project

import (
	"errors"
	"net/http"

	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the project module's repository, service, and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	projects := rg.Group("/projects")
	projects.POST("", m.create)
	projects.GET("", m.list)
	projects.GET("/:id", m.getBySlug)
	projects.POST("/:id/environments", m.addEnvironment)
	projects.GET("/:id/environments", m.listEnvironments)
	projects.POST("/:id/services", m.linkService)
	projects.GET("/:id/services", m.listServices)
	projects.DELETE("/:id/services/:serviceId", m.unlinkService)
	projects.POST("/:id/dependencies", m.addDependency)
	projects.GET("/:id/dependencies", m.listDependencies)
	projects.GET("/:id/dependents", m.listDependents)
	projects.DELETE("/:id/dependencies/:dependsOnId", m.removeDependency)
}

func (m *Module) create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	proj, err := m.service.Create(c.Request.Context(), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, proj)
}

func (m *Module) list(c *gin.Context) {
	projects, err := m.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, projects)
}

func (m *Module) getBySlug(c *gin.Context) {
	proj, err := m.service.GetBySlug(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, proj)
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

func (m *Module) linkService(c *gin.Context) {
	var req LinkServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := m.service.LinkService(c.Request.Context(), c.Param("id"), req); err != nil {
		m.respondServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (m *Module) listServices(c *gin.Context) {
	ids, err := m.service.ListServiceIDs(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, ids)
}

func (m *Module) unlinkService(c *gin.Context) {
	if err := m.service.UnlinkService(c.Request.Context(), c.Param("id"), c.Param("serviceId")); err != nil {
		m.respondServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
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

// listDependents shows what would be affected if this project were deleted.
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
