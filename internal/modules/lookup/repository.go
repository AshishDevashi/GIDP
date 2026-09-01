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

func (r *Repository) ListLifecycles(ctx context.Context) ([]store.Lifecycle, error) {
	return r.queries.ListLifecycles(ctx)
}

func (r *Repository) ListTiers(ctx context.Context) ([]store.Tier, error) {
	return r.queries.ListTiers(ctx)
}

func (r *Repository) ListServiceTypes(ctx context.Context) ([]store.ServiceType, error) {
	return r.queries.ListServiceTypes(ctx)
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
