package deploymentinstance

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateDeploymentInstanceParams) (store.Deploymentinstance, error) {
	return r.queries.CreateDeploymentInstance(ctx, arg)
}

func (r *Repository) GetLive(ctx context.Context) (store.Deploymentinstance, error) {
	return r.queries.GetLiveDeploymentInstance(ctx)
}

func (r *Repository) MarkProvisioned(ctx context.Context, arg store.MarkDeploymentInstanceProvisionedParams) (store.Deploymentinstance, error) {
	return r.queries.MarkDeploymentInstanceProvisioned(ctx, arg)
}

func (r *Repository) MarkStatus(ctx context.Context, arg store.MarkDeploymentInstanceStatusParams) (int64, error) {
	return r.queries.MarkDeploymentInstanceStatus(ctx, arg)
}

func (r *Repository) MarkTerminated(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.MarkDeploymentInstanceTerminated(ctx, id)
}
