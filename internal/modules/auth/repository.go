package auth

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries needed for registration and login.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) GetRoleByName(ctx context.Context, name string) (store.Role, error) {
	return r.queries.GetRoleByName(ctx, name)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (store.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (store.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *Repository) GetUserByID(ctx context.Context, id pgtype.UUID) (store.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *Repository) CreateUser(ctx context.Context, arg store.CreateUserParams) (store.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

func (r *Repository) UpdateLastLogin(ctx context.Context, id pgtype.UUID) error {
	return r.queries.UpdateLastLogin(ctx, id)
}
