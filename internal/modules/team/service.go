package team

import (
	"context"
	"errors"

	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrSlugTaken = errors.New("team slug is already in use")
	ErrNotFound  = errors.New("team not found")
	ErrInvalidID = errors.New("invalid id")
)

// Service contains the team module's business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateTeamRequest) (Response, error) {
	if _, err := s.repo.GetBySlug(ctx, req.Slug); err == nil {
		return Response{}, ErrSlugTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}

	team, err := s.repo.Create(ctx, store.CreateTeamParams{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(team), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	teams, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]Response, len(teams))
	for i, t := range teams {
		resp[i] = toResponse(t)
	}
	return resp, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Response, error) {
	team, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(team), nil
}

func (s *Service) AddMember(ctx context.Context, teamID string, req AddMemberRequest) (MemberResponse, error) {
	tID, err := uuidutil.Parse(teamID)
	if err != nil {
		return MemberResponse{}, ErrInvalidID
	}
	uID, err := uuidutil.Parse(req.UserID)
	if err != nil {
		return MemberResponse{}, ErrInvalidID
	}

	roleInTeam := req.RoleInTeam
	if roleInTeam == "" {
		roleInTeam = "member"
	}

	member, err := s.repo.AddMember(ctx, store.AddTeamMemberParams{
		TeamID:     tID,
		UserID:     uID,
		RoleInTeam: roleInTeam,
		IsPrimary:  req.IsPrimary,
	})
	if err != nil {
		return MemberResponse{}, err
	}

	return toMemberResponse(member), nil
}

func (s *Service) ListMembers(ctx context.Context, teamID string) ([]MemberResponse, error) {
	tID, err := uuidutil.Parse(teamID)
	if err != nil {
		return nil, ErrInvalidID
	}

	members, err := s.repo.ListMembers(ctx, tID)
	if err != nil {
		return nil, err
	}

	resp := make([]MemberResponse, len(members))
	for i, m := range members {
		resp[i] = toMemberResponse(m)
	}
	return resp, nil
}

func (s *Service) RemoveMember(ctx context.Context, teamID, userID string) error {
	tID, err := uuidutil.Parse(teamID)
	if err != nil {
		return ErrInvalidID
	}
	uID, err := uuidutil.Parse(userID)
	if err != nil {
		return ErrInvalidID
	}

	member, err := s.repo.GetActiveMember(ctx, tID, uID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return s.repo.RemoveMember(ctx, member.ID)
}

func toResponse(t store.Team) Response {
	return Response{
		ID:          uuidutil.String(t.ID),
		Name:        t.Name,
		Slug:        t.Slug,
		Description: t.Description.String,
		IsActive:    t.IsActive,
	}
}

func toMemberResponse(m store.TeamMember) MemberResponse {
	resp := MemberResponse{
		ID:         uuidutil.String(m.ID),
		TeamID:     uuidutil.String(m.TeamID),
		UserID:     uuidutil.String(m.UserID),
		RoleInTeam: m.RoleInTeam,
		IsPrimary:  m.IsPrimary,
	}
	if m.LeftAt.Valid {
		left := m.LeftAt.Time.String()
		resp.LeftAt = &left
	}
	return resp
}
