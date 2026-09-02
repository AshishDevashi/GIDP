package project

import (
	"context"
	"errors"

	"github.com/AshishDevashi/GIDP/internal/platform/pgnum"
	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DefaultProjectType    = "service"
	DefaultLifecycleID    = 2 // 'production', seeded in lookup tables migration
	DefaultDependencyType = "runtime"
)

var (
	ErrSlugTaken        = errors.New("project slug is already in use")
	ErrNotFound         = errors.New("project not found")
	ErrEnvironmentTaken = errors.New("environment is already registered for this project")
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidParent    = errors.New("project cannot be its own parent")
	ErrInvalidReference = errors.New("related resource does not exist")
	ErrProjectInUse     = errors.New("project is referenced by other resources")
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
	if err := s.validateParent(ctx, pgtype.UUID{}, parentProjectID); err != nil {
		return Response{}, err
	}

	projectType := defaultString(req.ProjectType, DefaultProjectType)
	lifecycleID := req.LifecycleID
	if lifecycleID == 0 {
		lifecycleID = DefaultLifecycleID
	}

	proj, err := s.repo.Create(ctx, store.CreateProjectParams{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     pgtext.From(req.Description),
		ProjectType:     projectType,
		Architecture:    pgtext.From(req.Architecture),
		OwnerTeamID:     ownerTeamID,
		TechLeadID:      techLeadID,
		LifecycleID:     lifecycleID,
		TierID:          pgnum.Int2From(req.TierID),
		DocsUrl:         pgtext.From(req.DocsURL),
		DashboardUrl:    pgtext.From(req.DashboardURL),
		RunbookUrl:      pgtext.From(req.RunbookURL),
		ParentProjectID: parentProjectID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Response{}, ErrSlugTaken
		}
		if isForeignKeyViolation(err) {
			return Response{}, ErrInvalidReference
		}
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

func (s *Service) Update(ctx context.Context, id string, req UpdateProjectRequest) (Response, error) {
	projectID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	current, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}

	if existing, err := s.repo.GetBySlug(ctx, req.Slug); err == nil {
		if existing.ID != projectID {
			return Response{}, ErrSlugTaken
		}
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
	if err := s.validateParent(ctx, projectID, parentProjectID); err != nil {
		return Response{}, err
	}

	projectType := defaultString(req.ProjectType, DefaultProjectType)
	lifecycleID := req.LifecycleID
	if lifecycleID == 0 {
		lifecycleID = DefaultLifecycleID
	}
	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	proj, err := s.repo.Update(ctx, store.UpdateProjectParams{
		ID:              projectID,
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     pgtext.From(req.Description),
		ProjectType:     projectType,
		Architecture:    pgtext.From(req.Architecture),
		OwnerTeamID:     ownerTeamID,
		TechLeadID:      techLeadID,
		LifecycleID:     lifecycleID,
		TierID:          pgnum.Int2From(req.TierID),
		DocsUrl:         pgtext.From(req.DocsURL),
		DashboardUrl:    pgtext.From(req.DashboardURL),
		RunbookUrl:      pgtext.From(req.RunbookURL),
		ParentProjectID: parentProjectID,
		IsActive:        isActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return Response{}, ErrSlugTaken
		}
		if isForeignKeyViolation(err) {
			return Response{}, ErrInvalidReference
		}
		return Response{}, err
	}

	return toResponse(proj), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	projectID, err := uuidutil.Parse(id)
	if err != nil {
		return ErrInvalidID
	}

	if _, err := s.repo.GetByID(ctx, projectID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if inUse, err := s.isProjectInUse(ctx, projectID); err != nil {
		return err
	} else if inUse {
		return ErrProjectInUse
	}

	rows, err := s.repo.Delete(ctx, projectID)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrProjectInUse
		}
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
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

// AddDependency records that projectID depends on req.DependsOnProjectID.
func (s *Service) AddDependency(ctx context.Context, projectID string, req AddDependencyRequest) (DependencyResponse, error) {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return DependencyResponse{}, ErrInvalidID
	}
	dependsOnID, err := uuidutil.Parse(req.DependsOnProjectID)
	if err != nil {
		return DependencyResponse{}, ErrInvalidID
	}

	dependencyType := defaultString(req.DependencyType, DefaultDependencyType)

	dep, err := s.repo.AddDependency(ctx, store.AddProjectDependencyParams{
		ProjectID:          pID,
		DependsOnProjectID: dependsOnID,
		DependencyType:     dependencyType,
	})
	if err != nil {
		return DependencyResponse{}, err
	}

	return toDependencyResponse(dep), nil
}

// ListDependencies returns the projects that projectID depends on.
func (s *Service) ListDependencies(ctx context.Context, projectID string) ([]DependencyResponse, error) {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return nil, ErrInvalidID
	}

	deps, err := s.repo.ListDependencies(ctx, pID)
	if err != nil {
		return nil, err
	}

	resp := make([]DependencyResponse, len(deps))
	for i, d := range deps {
		resp[i] = toDependencyResponse(d)
	}
	return resp, nil
}

// ListDependents returns the projects that depend on projectID — the impact list to show
// before deleting it.
func (s *Service) ListDependents(ctx context.Context, projectID string) ([]DependencyResponse, error) {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return nil, ErrInvalidID
	}

	deps, err := s.repo.ListDependents(ctx, pID)
	if err != nil {
		return nil, err
	}

	resp := make([]DependencyResponse, len(deps))
	for i, d := range deps {
		resp[i] = toDependencyResponse(d)
	}
	return resp, nil
}

func (s *Service) RemoveDependency(ctx context.Context, projectID, dependsOnProjectID string) error {
	pID, err := uuidutil.Parse(projectID)
	if err != nil {
		return ErrInvalidID
	}
	dependsOnID, err := uuidutil.Parse(dependsOnProjectID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.RemoveDependency(ctx, pID, dependsOnID)
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

func (s *Service) isProjectInUse(ctx context.Context, projectID pgtype.UUID) (bool, error) {
	checks := []func(context.Context, pgtype.UUID) (bool, error){
		func(ctx context.Context, id pgtype.UUID) (bool, error) {
			items, err := s.repo.ListChildren(ctx, id)
			return len(items) > 0, err
		},
		func(ctx context.Context, id pgtype.UUID) (bool, error) {
			items, err := s.repo.ListEnvironments(ctx, id)
			return len(items) > 0, err
		},
		func(ctx context.Context, id pgtype.UUID) (bool, error) {
			items, err := s.repo.ListDependencies(ctx, id)
			return len(items) > 0, err
		},
		func(ctx context.Context, id pgtype.UUID) (bool, error) {
			items, err := s.repo.ListDependents(ctx, id)
			return len(items) > 0, err
		},
		func(ctx context.Context, id pgtype.UUID) (bool, error) {
			items, err := s.repo.ListServices(ctx, id)
			return len(items) > 0, err
		},
		func(ctx context.Context, id pgtype.UUID) (bool, error) {
			items, err := s.repo.ListServicesByProject(ctx, id)
			return len(items) > 0, err
		},
	}

	for _, check := range checks {
		inUse, err := check(ctx, projectID)
		if err != nil || inUse {
			return inUse, err
		}
	}
	return false, nil
}

func (s *Service) validateParent(ctx context.Context, projectID, parentProjectID pgtype.UUID) error {
	if !parentProjectID.Valid {
		return nil
	}

	ancestorID := parentProjectID
	for ancestorID.Valid {
		if projectID.Valid && ancestorID.Bytes == projectID.Bytes {
			return ErrInvalidParent
		}

		ancestor, err := s.repo.GetByID(ctx, ancestorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidReference
			}
			return err
		}
		ancestorID = ancestor.ParentProjectID
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
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
		LifecycleID:     p.LifecycleID,
		TierID:          pgnum.Int2To(p.TierID),
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

func toDependencyResponse(d store.ProjectDependency) DependencyResponse {
	return DependencyResponse{
		ID:                 uuidutil.String(d.ID),
		ProjectID:          uuidutil.String(d.ProjectID),
		DependsOnProjectID: uuidutil.String(d.DependsOnProjectID),
		DependencyType:     d.DependencyType,
	}
}
