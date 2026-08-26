package team

import (
	"errors"
	"net/http"

	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the team module's repository, service, and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	teams := rg.Group("/teams")
	teams.POST("", m.create)
	teams.GET("", m.list)
	teams.GET("/:id", m.getBySlug)
	teams.POST("/:id/members", m.addMember)
	teams.GET("/:id/members", m.listMembers)
	teams.DELETE("/:id/members/:userId", m.removeMember)
}

func (m *Module) create(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	team, err := m.service.Create(c.Request.Context(), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, team)
}

func (m *Module) list(c *gin.Context) {
	teams, err := m.service.List(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, teams)
}

func (m *Module) getBySlug(c *gin.Context) {
	team, err := m.service.GetBySlug(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, team)
}

func (m *Module) addMember(c *gin.Context) {
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	member, err := m.service.AddMember(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, member)
}

func (m *Module) listMembers(c *gin.Context) {
	members, err := m.service.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, members)
}

func (m *Module) removeMember(c *gin.Context) {
	err := m.service.RemoveMember(c.Request.Context(), c.Param("id"), c.Param("userId"))
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSlugTaken):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidID):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
