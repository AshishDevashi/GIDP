package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/AshishDevashi/GIDP/internal/config"
	"github.com/AshishDevashi/GIDP/internal/modules/auth"
	"github.com/AshishDevashi/GIDP/internal/modules/deployment"
	"github.com/AshishDevashi/GIDP/internal/modules/health"
	"github.com/AshishDevashi/GIDP/internal/modules/lookup"
	"github.com/AshishDevashi/GIDP/internal/modules/project"
	"github.com/AshishDevashi/GIDP/internal/modules/service"
	"github.com/AshishDevashi/GIDP/internal/modules/team"
	"github.com/AshishDevashi/GIDP/internal/modules/user"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server wires together the HTTP router and all registered modules.
type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *pgxpool.Pool
	router *gin.Engine
}

// New constructs a Server and registers all module routes.
func New(cfg *config.Config, log *slog.Logger, db *pgxpool.Pool) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{
		cfg:    cfg,
		log:    log,
		db:     db,
		router: gin.New(),
	}

	s.router.Use(gin.Recovery())
	s.registerModules()

	return s
}

func (s *Server) registerModules() {
	health.NewModule().RegisterRoutes(s.router)

	v1 := s.router.Group("/api/v1")

	authModule := auth.NewModule(s.db, s.cfg.JWTSecret, s.cfg.JWTTTL)
	authModule.RegisterRoutes(v1)

	protected := v1.Group("")
	protected.Use(authModule.RequireAuth())
	user.NewModule(s.db).RegisterRoutes(protected)
	team.NewModule(s.db).RegisterRoutes(protected)
	project.NewModule(s.db).RegisterRoutes(protected)
	service.NewModule(s.db).RegisterRoutes(protected)
	deployment.NewModule(s.db).RegisterRoutes(protected)
	lookup.NewModule(s.db).RegisterRoutes(protected)
}

// Router exposes the underlying gin engine, e.g. for route introspection tooling.
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Run starts the HTTP server and blocks until the context is cancelled or an error occurs.
func (s *Server) Run(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("starting server", "port", s.cfg.Port, "env", s.cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}
