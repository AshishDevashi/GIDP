package repo

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the repo module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateRepoParams) (store.Repo, error) {
	return r.queries.CreateRepo(ctx, arg)
}

func (r *Repository) GetTemplateBySlug(ctx context.Context, slug string) (store.RepoTemplate, error) {
	return r.queries.GetRepoTemplateBySlug(ctx, slug)
}

func (r *Repository) MarkCreating(ctx context.Context, id pgtype.UUID) (store.Repo, error) {
	return r.queries.MarkRepoCreating(ctx, id)
}

func (r *Repository) Activate(ctx context.Context, arg store.ActivateRepoParams) (store.Repo, error) {
	return r.queries.ActivateRepo(ctx, arg)
}

func (r *Repository) Fail(ctx context.Context, arg store.FailRepoParams) (store.Repo, error) {
	return r.queries.FailRepo(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Repo, error) {
	return r.queries.GetRepoByID(ctx, id)
}

func (r *Repository) List(ctx context.Context) ([]store.Repo, error) {
	return r.queries.ListRepos(ctx)
}

func (r *Repository) Update(ctx context.Context, arg store.UpdateRepoParams) (store.Repo, error) {
	return r.queries.UpdateRepo(ctx, arg)
}

func (r *Repository) Delete(ctx context.Context, id pgtype.UUID) (int64, error) {
	return r.queries.DeleteRepo(ctx, id)
}
