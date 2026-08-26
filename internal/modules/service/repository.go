package service

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the service module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateServiceParams) (store.Service, error) {
	return r.queries.CreateService(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Service, error) {
	return r.queries.GetServiceByID(ctx, id)
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (store.Service, error) {
	return r.queries.GetServiceBySlug(ctx, slug)
}

func (r *Repository) List(ctx context.Context) ([]store.Service, error) {
	return r.queries.ListServices(ctx)
}

func (r *Repository) ListByProject(ctx context.Context, projectID pgtype.UUID) ([]store.Service, error) {
	return r.queries.ListServicesByProject(ctx, projectID)
}

func (r *Repository) AddEnvironment(ctx context.Context, arg store.AddServiceEnvironmentParams) (store.ServiceEnvironment, error) {
	return r.queries.AddServiceEnvironment(ctx, arg)
}

func (r *Repository) GetEnvironment(ctx context.Context, serviceID pgtype.UUID, environment string) (store.ServiceEnvironment, error) {
	return r.queries.GetServiceEnvironment(ctx, store.GetServiceEnvironmentParams{ServiceID: serviceID, Environment: environment})
}

func (r *Repository) ListEnvironments(ctx context.Context, serviceID pgtype.UUID) ([]store.ServiceEnvironment, error) {
	return r.queries.ListServiceEnvironments(ctx, serviceID)
}

func (r *Repository) AddDependency(ctx context.Context, arg store.AddServiceDependencyParams) (store.ServiceDependency, error) {
	return r.queries.AddServiceDependency(ctx, arg)
}

func (r *Repository) ListDependencies(ctx context.Context, serviceID pgtype.UUID) ([]store.ServiceDependency, error) {
	return r.queries.ListServiceDependencies(ctx, serviceID)
}

func (r *Repository) ListDependents(ctx context.Context, serviceID pgtype.UUID) ([]store.ServiceDependency, error) {
	return r.queries.ListDependentServices(ctx, serviceID)
}

func (r *Repository) RemoveDependency(ctx context.Context, serviceID, dependsOnServiceID pgtype.UUID) error {
	return r.queries.RemoveServiceDependency(ctx, store.RemoveServiceDependencyParams{
		ServiceID:          serviceID,
		DependsOnServiceID: dependsOnServiceID,
	})
}

func (r *Repository) AddTag(ctx context.Context, serviceID pgtype.UUID, tag string) error {
	return r.queries.AddServiceTag(ctx, store.AddServiceTagParams{ServiceID: serviceID, Tag: tag})
}

func (r *Repository) RemoveTag(ctx context.Context, serviceID pgtype.UUID, tag string) error {
	return r.queries.RemoveServiceTag(ctx, store.RemoveServiceTagParams{ServiceID: serviceID, Tag: tag})
}

func (r *Repository) ListTags(ctx context.Context, serviceID pgtype.UUID) ([]store.ServiceTag, error) {
	return r.queries.ListServiceTags(ctx, serviceID)
}
