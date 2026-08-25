package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/token"
	"github.com/AshishDevashi/GIDP/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the auth module's repository, service, and HTTP routes together.
type Module struct {
	service   *Service
	jwtSecret string
}

func NewModule(pool *pgxpool.Pool, jwtSecret string, jwtTTL time.Duration) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo, jwtSecret, jwtTTL)
	return &Module{service: svc, jwtSecret: jwtSecret}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/auth")
	group.POST("/register", m.register)
	group.POST("/login", m.login)
	group.GET("/me", m.RequireAuth(), m.me)
}

func (m *Module) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := m.service.Register(c.Request.Context(), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, resp)
}

func (m *Module) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := m.service.Login(c.Request.Context(), req)
	if err != nil {
		m.respondServiceError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, resp)
}

func (m *Module) me(c *gin.Context) {
	userID := c.GetString(ContextUserIDKey)

	user, err := m.service.Me(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.JSON(c, http.StatusOK, user)
}

func (m *Module) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken), errors.Is(err, ErrUsernameTaken):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(c, http.StatusUnauthorized, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}

// Context keys populated by RequireAuth for downstream handlers.
const (
	ContextUserIDKey   = "auth_user_id"
	ContextUsernameKey = "auth_username"
	ContextRoleIDKey   = "auth_role_id"
)

// RequireAuth validates the Bearer JWT on the request and stores its claims in the context.
func (m *Module) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, "missing or malformed authorization header")
			c.Abort()
			return
		}

		claims, err := token.Parse(m.jwtSecret, parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextRoleIDKey, claims.RoleID)
		c.Next()
	}
}
