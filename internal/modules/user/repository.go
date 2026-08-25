package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wolf-platform/wolf-platform/internal/modules/user/db"
)

// Repository wraps sqlc-generated queries for the user module.
type Repository struct {
	queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: db.New(pool)}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (db.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *Repository) List(ctx context.Context) ([]db.User, error) {
	return r.queries.ListUsers(ctx)
}
