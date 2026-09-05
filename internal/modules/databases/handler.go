package databases

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/AshishDevashi/GIDP/internal/modules/auth"
	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the databases module service and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool, log *slog.Logger) *Module {
	repo := NewRepository(pool)
	client := NewPostgresClient()
	resolver := NewSecretResolver()
	return &Module{service: NewService(repo, client, resolver, log)}
}

func NewModuleWithDeps(repo *Repository, client PostgresClient, resolver *SecretResolver, log *slog.Logger) *Module {
	return &Module{service: NewService(repo, client, resolver, log)}
}

// Service returns the underlying database service (e.g. for cascade hooks).
func (m *Module) Service() *Service {
	return m.service
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	dbs := rg.Group("/databases")
	dbs.POST("", m.create)
	dbs.GET("", m.list)
	dbs.GET("/quota", m.getQuota)
	dbs.GET("/:id", m.getByID)
	dbs.GET("/:id/connection-string", m.getConnectionString)
	dbs.GET("/:id/connect", m.getConnectionString)
	dbs.DELETE("/:id", m.delete)
}

func (m *Module) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	db, err := m.service.CreateDatabase(c.Request.Context(), req, c.GetString(auth.ContextUserIDKey))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, db)
}

func (m *Module) list(c *gin.Context) {
	instanceID := c.Query("db_instance_id")
	var databases []Response
	var err error

	if instanceID != "" {
		databases, err = m.service.ListByInstanceID(c.Request.Context(), instanceID)
	} else {
		databases, err = m.service.List(c.Request.Context())
	}

	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, databases)
}

func (m *Module) getByID(c *gin.Context) {
	db, err := m.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, db)
}

func (m *Module) getConnectionString(c *gin.Context) {
	conn, err := m.service.GetConnectionString(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, conn)
}

func (m *Module) delete(c *gin.Context) {
	if err := m.service.DeleteDatabase(c.Request.Context(), c.Param("id")); err != nil {
		m.respondServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (m *Module) getQuota(c *gin.Context) {
	instanceID := c.Query("db_instance_id")
	quota, err := m.service.GetQuota(c.Request.Context(), instanceID)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, quota)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidID), errors.Is(err, ErrInvalidName), errors.Is(err, ErrInvalidSize):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAlreadyExists), errors.Is(err, ErrQuotaExceeded):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNoActiveDBInstance):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInstanceNotReady):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
