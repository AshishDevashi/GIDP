package databases

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for managed databases.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateManagedDatabaseParams) (store.ManagedDatabase, error) {
	return r.queries.CreateManagedDatabase(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.ManagedDatabase, error) {
	return r.queries.GetManagedDatabaseByID(ctx, id)
}

func (r *Repository) GetByName(ctx context.Context, instanceID pgtype.UUID, name string) (store.ManagedDatabase, error) {
	return r.queries.GetManagedDatabaseByName(ctx, store.GetManagedDatabaseByNameParams{
		DbInstanceID: instanceID,
		Name:         name,
	})
}

func (r *Repository) List(ctx context.Context) ([]store.ManagedDatabase, error) {
	return r.queries.ListManagedDatabases(ctx)
}

func (r *Repository) ListByInstanceID(ctx context.Context, instanceID pgtype.UUID) ([]store.ManagedDatabase, error) {
	return r.queries.ListManagedDatabasesByInstanceID(ctx, instanceID)
}

func (r *Repository) ListActiveByInstanceID(ctx context.Context, instanceID pgtype.UUID) ([]store.ManagedDatabase, error) {
	return r.queries.ListActiveManagedDatabasesByInstanceID(ctx, instanceID)
}

func (r *Repository) GetTotalAllocatedMB(ctx context.Context) (int64, error) {
	return r.queries.GetTotalAllocatedMB(ctx)
}

func (r *Repository) GetTotalAllocatedMBByInstanceID(ctx context.Context, instanceID pgtype.UUID) (int64, error) {
	return r.queries.GetTotalAllocatedMBByInstanceID(ctx, instanceID)
}

func (r *Repository) MarkStatus(ctx context.Context, id pgtype.UUID, status string) (int64, error) {
	return r.queries.MarkManagedDatabaseStatus(ctx, store.MarkManagedDatabaseStatusParams{
		ID:     id,
		Status: status,
	})
}

func (r *Repository) SoftDelete(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.SoftDeleteManagedDatabase(ctx, id)
}

func (r *Repository) SoftDeleteByInstanceID(ctx context.Context, instanceID pgtype.UUID) (int64, error) {
	return r.queries.SoftDeleteManagedDatabasesByInstanceID(ctx, instanceID)
}

func (r *Repository) ListActiveDBInstances(ctx context.Context) ([]store.DbInstance, error) {
	return r.queries.ListDBInstances(ctx)
}

func (r *Repository) GetDBInstanceByID(ctx context.Context, id pgtype.UUID) (store.DbInstance, error) {
	return r.queries.GetDBInstanceByID(ctx, id)
}
