package project

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the project module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateProjectParams) (store.Project, error) {
	return r.queries.CreateProject(ctx, arg)
}

func (r *Repository) Update(ctx context.Context, arg store.UpdateProjectParams) (store.Project, error) {
	return r.queries.UpdateProject(ctx, arg)
}

func (r *Repository) Delete(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.DeleteProject(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Project, error) {
	return r.queries.GetProjectByID(ctx, id)
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (store.Project, error) {
	return r.queries.GetProjectBySlug(ctx, slug)
}

func (r *Repository) List(ctx context.Context) ([]store.Project, error) {
	return r.queries.ListProjects(ctx)
}

func (r *Repository) ListChildren(ctx context.Context, parentID pgtype.UUID) ([]store.Project, error) {
	return r.queries.ListChildProjects(ctx, parentID)
}

func (r *Repository) AddEnvironment(ctx context.Context, arg store.AddProjectEnvironmentParams) (store.ProjectEnvironment, error) {
	return r.queries.AddProjectEnvironment(ctx, arg)
}

func (r *Repository) GetEnvironment(ctx context.Context, projectID pgtype.UUID, environment string) (store.ProjectEnvironment, error) {
	return r.queries.GetProjectEnvironment(ctx, store.GetProjectEnvironmentParams{ProjectID: projectID, Environment: environment})
}

func (r *Repository) ListEnvironments(ctx context.Context, projectID pgtype.UUID) ([]store.ProjectEnvironment, error) {
	return r.queries.ListProjectEnvironments(ctx, projectID)
}

func (r *Repository) LinkService(ctx context.Context, projectID, serviceID pgtype.UUID) error {
	return r.queries.LinkProjectService(ctx, store.LinkProjectServiceParams{ProjectID: projectID, ServiceID: serviceID})
}

func (r *Repository) UnlinkService(ctx context.Context, projectID, serviceID pgtype.UUID) error {
	return r.queries.UnlinkProjectService(ctx, store.UnlinkProjectServiceParams{ProjectID: projectID, ServiceID: serviceID})
}

func (r *Repository) ListServices(ctx context.Context, projectID pgtype.UUID) ([]store.ProjectService, error) {
	return r.queries.ListProjectServices(ctx, projectID)
}

func (r *Repository) ListServicesByProject(ctx context.Context, projectID pgtype.UUID) ([]store.Service, error) {
	return r.queries.ListServicesByProject(ctx, projectID)
}

func (r *Repository) AddDependency(ctx context.Context, arg store.AddProjectDependencyParams) (store.ProjectDependency, error) {
	return r.queries.AddProjectDependency(ctx, arg)
}

func (r *Repository) ListDependencies(ctx context.Context, projectID pgtype.UUID) ([]store.ProjectDependency, error) {
	return r.queries.ListProjectDependencies(ctx, projectID)
}

func (r *Repository) ListDependents(ctx context.Context, projectID pgtype.UUID) ([]store.ProjectDependency, error) {
	return r.queries.ListDependentProjects(ctx, projectID)
}

func (r *Repository) RemoveDependency(ctx context.Context, projectID, dependsOnProjectID pgtype.UUID) error {
	return r.queries.RemoveProjectDependency(ctx, store.RemoveProjectDependencyParams{
		ProjectID:          projectID,
		DependsOnProjectID: dependsOnProjectID,
	})
}
