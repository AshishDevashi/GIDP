package team

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc-generated queries for the team module.
type Repository struct {
	queries *store.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: store.New(pool)}
}

func (r *Repository) Create(ctx context.Context, arg store.CreateTeamParams) (store.Team, error) {
	return r.queries.CreateTeam(ctx, arg)
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (store.Team, error) {
	return r.queries.GetTeamBySlug(ctx, slug)
}

func (r *Repository) GetByID(ctx context.Context, id pgtype.UUID) (store.Team, error) {
	return r.queries.GetTeamByID(ctx, id)
}

func (r *Repository) List(ctx context.Context) ([]store.Team, error) {
	return r.queries.ListTeams(ctx)
}

func (r *Repository) AddMember(ctx context.Context, arg store.AddTeamMemberParams) (store.TeamMember, error) {
	return r.queries.AddTeamMember(ctx, arg)
}

func (r *Repository) ListMembers(ctx context.Context, teamID pgtype.UUID) ([]store.TeamMember, error) {
	return r.queries.ListTeamMembers(ctx, teamID)
}

func (r *Repository) GetMember(ctx context.Context, teamID, userID pgtype.UUID) (store.TeamMember, error) {
	return r.queries.GetTeamMember(ctx, store.GetTeamMemberParams{TeamID: teamID, UserID: userID})
}

func (r *Repository) RemoveMember(ctx context.Context, teamID, userID pgtype.UUID) error {
	return r.queries.RemoveTeamMember(ctx, store.RemoveTeamMemberParams{TeamID: teamID, UserID: userID})
}
