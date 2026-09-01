package repo

import (
	"errors"
	"net/http"

	"github.com/AshishDevashi/GIDP/internal/modules/auth"
	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the repo module's service and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool, githubToken string) *Module {
	repo := NewRepository(pool)
	return &Module{service: NewService(repo, githubToken)}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	repos := rg.Group("/repos")
	repos.POST("", m.create)
	repos.GET("", m.list)
	repos.GET("/:id", m.getByID)
	repos.PUT("/:id", m.update)
	repos.DELETE("/:id", m.delete)
}

func (m *Module) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	repository, err := m.service.Create(
		c.Request.Context(),
		req,
		c.GetString(auth.ContextUserIDKey),
		c.GetString(auth.ContextUsernameKey),
	)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, repository)
}

func (m *Module) list(c *gin.Context) {
	repositories, err := m.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list repositories")
		return
	}
	response.JSON(c, http.StatusOK, repositories)
}

func (m *Module) getByID(c *gin.Context) {
	repository, err := m.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, repository)
}

func (m *Module) update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	repository, err := m.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, repository)
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
	case errors.Is(err, ErrRepoTaken):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrUnauthorized):
		response.Error(c, http.StatusBadGateway, err.Error())
	case errors.Is(err, ErrOrganization):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrRepositoryInvalid):
		response.Error(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrNotConfigured), errors.Is(err, ErrGitHubUnavailable):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, "repository operation failed")
	}
}
