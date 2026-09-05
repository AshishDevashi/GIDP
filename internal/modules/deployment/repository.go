package deployment

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

func (r *Repository) Create(ctx context.Context, arg store.CreateDeploymentParams) (store.Deployment, error) {
	return r.queries.CreateDeployment(ctx, arg)
}

func (r *Repository) CreateRevision(ctx context.Context, arg store.CreateDeploymentRevisionParams) (store.Deploymentrevision, error) {
	return r.queries.CreateDeploymentRevision(ctx, arg)
}

func (r *Repository) CountActiveByInstanceID(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.CountActiveDeploymentsByInstanceID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Deployment, error) {
	return r.queries.GetDeploymentByID(ctx, id)
}

func (r *Repository) List(ctx context.Context) ([]store.Deployment, error) {
	return r.queries.ListDeployments(ctx)
}

func (r *Repository) LatestRevision(ctx context.Context, id pgtype.UUID) (store.Deploymentrevision, error) {
	return r.queries.GetLatestDeploymentRevision(ctx, id)
}

func (r *Repository) MarkDeploying(ctx context.Context, id pgtype.UUID) (store.Deployment, error) {
	return r.queries.MarkDeploymentDeploying(ctx, id)
}

func (r *Repository) MarkStatus(ctx context.Context, arg store.MarkDeploymentStatusParams) (int64, error) {
	return r.queries.MarkDeploymentStatus(ctx, arg)
}

func (r *Repository) Delete(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.DeleteDeployment(ctx, id)
}

func (r *Repository) GetActiveDeploymentInstance(ctx context.Context) (store.Deploymentinstance, error) {
	return r.queries.GetLiveDeploymentInstance(ctx)
}

func (r *Repository) GetRegistryByID(ctx context.Context, id pgtype.UUID) (store.Registry, error) {
	return r.queries.GetRegistryByID(ctx, id)
}

func (r *Repository) GetRepoByID(ctx context.Context, id pgtype.UUID) (store.Repo, error) {
	return r.queries.GetRepoByID(ctx, id)
}
