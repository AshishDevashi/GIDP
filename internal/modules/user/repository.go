package user

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the user module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *Repository) List(ctx context.Context) ([]store.User, error) {
	return r.queries.ListUsers(ctx)
}
