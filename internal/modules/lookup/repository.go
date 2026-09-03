package lookup

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the read-only lookup module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) ListRepoProviders(ctx context.Context) ([]store.RepoProvider, error) {
	return r.queries.ListRepoProviders(ctx)
}

func (r *Repository) ListLanguages(ctx context.Context) ([]store.Language, error) {
	return r.queries.ListLanguages(ctx)
}

func (r *Repository) ListRepoTemplates(ctx context.Context) ([]store.RepoTemplate, error) {
	return r.queries.ListRepoTemplates(ctx)
}

func (r *Repository) ListRegistryProviders(ctx context.Context) ([]store.RegistryProvider, error) {
	return r.queries.ListRegistryProviders(ctx)
}
