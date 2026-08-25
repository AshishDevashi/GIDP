package user

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wolf-platform/wolf-platform/pkg/response"
)

// Module wires the user module's repository, service, and HTTP routes together.
type Module struct {
	service *Service
}

func NewModule(pool *pgxpool.Pool) *Module {
	repo := NewRepository(pool)
	svc := NewService(repo)
	return &Module{service: svc}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/users", m.list)
}

func (m *Module) list(w http.ResponseWriter, r *http.Request) {
	users, err := m.service.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}
