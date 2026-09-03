package registry

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the registry module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateRegistryParams) (store.Registry, error) {
	return r.queries.CreateRegistry(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Registry, error) {
	return r.queries.GetRegistryByID(ctx, id)
}

func (r *Repository) GetByName(ctx context.Context, arg store.GetRegistryByNameParams) (store.Registry, error) {
	return r.queries.GetRegistryByName(ctx, arg)
}

func (r *Repository) List(ctx context.Context) ([]store.Registry, error) {
	return r.queries.ListRegistries(ctx)
}

func (r *Repository) Update(ctx context.Context, arg store.UpdateRegistryParams) (store.Registry, error) {
	return r.queries.UpdateRegistry(ctx, arg)
}

func (r *Repository) SoftDelete(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.SoftDeleteRegistry(ctx, id)
}
