package dbinstance

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the db instance module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateDBInstanceParams) (store.DbInstance, error) {
	return r.queries.CreateDBInstance(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.DbInstance, error) {
	return r.queries.GetDBInstanceByID(ctx, id)
}

func (r *Repository) GetByName(ctx context.Context, name string) (store.DbInstance, error) {
	return r.queries.GetDBInstanceByName(ctx, name)
}

func (r *Repository) List(ctx context.Context) ([]store.DbInstance, error) {
	return r.queries.ListDBInstances(ctx)
}

func (r *Repository) CountActive(ctx context.Context) (int64, error) {
	return r.queries.CountActiveDBInstances(ctx)
}

func (r *Repository) MarkProvisioned(ctx context.Context, arg store.MarkDBInstanceProvisionedParams) (store.DbInstance, error) {
	return r.queries.MarkDBInstanceProvisioned(ctx, arg)
}

func (r *Repository) MarkStatus(ctx context.Context, arg store.MarkDBInstanceStatusParams) (int64, error) {
	return r.queries.MarkDBInstanceStatus(ctx, arg)
}

func (r *Repository) MarkContainerStatus(ctx context.Context, arg store.MarkDBInstanceContainerStatusParams) (int64, error) {
	return r.queries.MarkDBInstanceContainerStatus(ctx, arg)
}

func (r *Repository) SoftDelete(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.SoftDeleteDBInstance(ctx, id)
}
