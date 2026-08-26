package deployment

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the deployment module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateDeploymentParams) (store.Deployment, error) {
	return r.queries.CreateDeployment(ctx, arg)
}

func (r *Repository) UpdateStatus(ctx context.Context, arg store.UpdateDeploymentStatusParams) (store.Deployment, error) {
	return r.queries.UpdateDeploymentStatus(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Deployment, error) {
	return r.queries.GetDeploymentByID(ctx, id)
}

func (r *Repository) ListByService(ctx context.Context, serviceID pgtype.UUID) ([]store.Deployment, error) {
	return r.queries.ListDeploymentsByService(ctx, serviceID)
}

func (r *Repository) ListByServiceEnvironment(ctx context.Context, serviceID pgtype.UUID, environment string) ([]store.Deployment, error) {
	return r.queries.ListDeploymentsByServiceEnvironment(ctx, store.ListDeploymentsByServiceEnvironmentParams{
		ServiceID:   serviceID,
		Environment: environment,
	})
}

func (r *Repository) GetCurrent(ctx context.Context, serviceID pgtype.UUID, environment string) (store.Deployment, error) {
	return r.queries.GetCurrentDeployment(ctx, store.GetCurrentDeploymentParams{
		ServiceID:   serviceID,
		Environment: environment,
	})
}
