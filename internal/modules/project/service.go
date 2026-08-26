package project

import (
	"context"
	"errors"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DefaultProjectType   = "service"
	DefaultLifecycle     = "production"
	DefaultDefaultBranch = "main"
)

var (
	ErrSlugTaken        = errors.New("project slug is already in use")
	ErrNotFound         = errors.New("project not found")
	ErrEnvironmentTaken = errors.New("environment is already registered for this project")
	ErrInvalidID        = errors.New("invalid id")
)

// Service contains the project module's business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateProjectRequest) (Response, error) {
	if _, err := s.repo.GetBySlug(ctx, req.Slug); err == nil {
		return Response{}, ErrSlugTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}

	ownerTeamID, err := uuidutil.Parse(req.OwnerTeamID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	techLeadID, err := optionalUUID(req.TechLeadID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	parentProjectID, err := optionalUUID(req.ParentProjectID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	projectType := defaultString(req.ProjectType, DefaultProjectType)
	lifecycle := defaultString(req.Lifecycle, DefaultLifecycle)
	defaultBranch := defaultString(req.DefaultBranch, DefaultDefaultBranch)

	proj, err := s.repo.Create(ctx, store.CreateProjectParams{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     pgtext.From(req.Description),
		ProjectType:     projectType,
		Architecture:    pgtext.From(req.Architecture),
		OwnerTeamID:     ownerTeamID,
		TechLeadID:      techLeadID,
		RepoUrl:         pgtext.From(req.RepoURL),
		RepoProvider:    pgtext.From(req.RepoProvider),
		DefaultBranch:   defaultBranch,
		CiPipelineUrl:   pgtext.From(req.CIPipelineURL),
		GitopsPath:      pgtext.From(req.GitopsPath),
		Lifecycle:       lifecycle,
		Tier:            pgtext.From(req.Tier),
		Language:        pgtext.From(req.Language),
		Framework:       pgtext.From(req.Framework),
		DocsUrl:         pgtext.From(req.DocsURL),
		DashboardUrl:    pgtext.From(req.DashboardURL),
		RunbookUrl:      pgtext.From(req.RunbookURL),
		ParentProjectID: parentProjectID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(proj), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	projects, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]Response, len(projects))
	for i, p := range projects {
		resp[i] = toResponse(p)
	}
	return resp, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Response, error) {
	proj, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(proj), nil
}

func (s *Service) AddEnvironment(ctx context.Context, projectID string, req AddEnvironmentRequest) (EnvironmentResponse, error) {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return EnvironmentResponse{}, ErrInvalidID
	}

	if _, err := s.repo.GetEnvironment(ctx, pID, req.Environment); err == nil {
		return EnvironmentResponse{}, ErrEnvironmentTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return EnvironmentResponse{}, err
	}

	replicas := req.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	env, err := s.repo.AddEnvironment(ctx, store.AddProjectEnvironmentParams{
		ProjectID:   pID,
		Environment: req.Environment,
		ClusterName: pgtext.From(req.ClusterName),
		Namespace:   pgtext.From(req.Namespace),
		Url:         pgtext.From(req.URL),
		Replicas:    replicas,
	})
	if err != nil {
		return EnvironmentResponse{}, err
	}

	return toEnvironmentResponse(env), nil
}

func (s *Service) ListEnvironments(ctx context.Context, projectID string) ([]EnvironmentResponse, error) {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return nil, ErrInvalidID
	}

	envs, err := s.repo.ListEnvironments(ctx, pID)
	if err != nil {
		return nil, err
	}

	resp := make([]EnvironmentResponse, len(envs))
	for i, e := range envs {
		resp[i] = toEnvironmentResponse(e)
	}
	return resp, nil
}

func (s *Service) LinkService(ctx context.Context, projectID string, req LinkServiceRequest) error {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return ErrInvalidID
	}
	sID, err := uuidutil.Parse(req.ServiceID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.LinkService(ctx, pID, sID)
}

func (s *Service) UnlinkService(ctx context.Context, projectID, serviceID string) error {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return ErrInvalidID
	}
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.UnlinkService(ctx, pID, sID)
}

func (s *Service) ListServiceIDs(ctx context.Context, projectID string) ([]string, error) {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return nil, ErrInvalidID
	}

	links, err := s.repo.ListServices(ctx, pID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(links))
	for i, l := range links {
		ids[i] = uuidutil.String(l.ServiceID)
	}
	return ids, nil
}

func optionalUUID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	return uuidutil.Parse(s)
}

func defaultString(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func toResponse(p store.Project) Response {
	return Response{
		ID:              uuidutil.String(p.ID),
		Name:            p.Name,
		Slug:            p.Slug,
		Description:     pgtext.To(p.Description),
		ProjectType:     p.ProjectType,
		Architecture:    pgtext.To(p.Architecture),
		OwnerTeamID:     uuidutil.String(p.OwnerTeamID),
		TechLeadID:      uuidutil.String(p.TechLeadID),
		RepoURL:         pgtext.To(p.RepoUrl),
		RepoProvider:    pgtext.To(p.RepoProvider),
		DefaultBranch:   p.DefaultBranch,
		CIPipelineURL:   pgtext.To(p.CiPipelineUrl),
		GitopsPath:      pgtext.To(p.GitopsPath),
		Lifecycle:       p.Lifecycle,
		Tier:            pgtext.To(p.Tier),
		Language:        pgtext.To(p.Language),
		Framework:       pgtext.To(p.Framework),
		DocsURL:         pgtext.To(p.DocsUrl),
		DashboardURL:    pgtext.To(p.DashboardUrl),
		RunbookURL:      pgtext.To(p.RunbookUrl),
		ParentProjectID: uuidutil.String(p.ParentProjectID),
		IsActive:        p.IsActive,
	}
}

func toEnvironmentResponse(e store.ProjectEnvironment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:          uuidutil.String(e.ID),
		ProjectID:   uuidutil.String(e.ProjectID),
		Environment: e.Environment,
		ClusterName: pgtext.To(e.ClusterName),
		Namespace:   pgtext.To(e.Namespace),
		URL:         pgtext.To(e.Url),
		Replicas:    e.Replicas,
	}
}
