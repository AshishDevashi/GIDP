package service

import (
	"context"
	"errors"

	"github.com/AshishDevashi/GIDP/internal/platform/pgnum"
	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DefaultServiceTypeID   = 1 // 'backend'
	DefaultLifecycleID     = 2 // 'production'
	DefaultDefaultBranch   = "main"
	DefaultDockerfilePath  = "Dockerfile"
	DefaultK8sResourceKind = "Deployment"
	DefaultHealthCheckPath = "/healthz"
	DefaultDependencyType  = "sync_api"
)

var (
	ErrSlugTaken        = errors.New("service slug is already in use")
	ErrNotFound         = errors.New("service not found")
	ErrEnvironmentTaken = errors.New("environment is already registered for this service")
	ErrInvalidID        = errors.New("invalid id")
)

// Service contains the service module's business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateServiceRequest) (Response, error) {
	if _, err := s.repo.GetBySlug(ctx, req.Slug); err == nil {
		return Response{}, ErrSlugTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}

	ownerTeamID, err := uuidutil.Parse(req.OwnerTeamID)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	projectID, err := optionalUUID(req.ProjectID)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	techLeadID, err := optionalUUID(req.TechLeadID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	serviceTypeID := req.ServiceTypeID
	if serviceTypeID == 0 {
		serviceTypeID = DefaultServiceTypeID
	}
	lifecycleID := req.LifecycleID
	if lifecycleID == 0 {
		lifecycleID = DefaultLifecycleID
	}

	svc, err := s.repo.Create(ctx, store.CreateServiceParams{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     pgtext.From(req.Description),
		ServiceTypeID:   serviceTypeID,
		LifecycleID:     pgnum.Int2From(lifecycleID),
		TierID:          pgnum.Int2From(req.TierID),
		ProjectID:       projectID,
		OwnerTeamID:     ownerTeamID,
		TechLeadID:      techLeadID,
		RepoUrl:         req.RepoURL,
		RepoProviderID:  pgnum.Int2From(req.RepoProviderID),
		DefaultBranch:   defaultString(req.DefaultBranch, DefaultDefaultBranch),
		LanguageID:      pgnum.Int2From(req.LanguageID),
		Framework:       pgtext.From(req.Framework),
		DockerfilePath:  defaultString(req.DockerfilePath, DefaultDockerfilePath),
		RegistryImage:   pgtext.From(req.RegistryImage),
		CiPipelineUrl:   pgtext.From(req.CIPipelineURL),
		GitopsRepoPath:  pgtext.From(req.GitopsRepoPath),
		K8sResourceKind: defaultString(req.K8sResourceKind, DefaultK8sResourceKind),
		Port:            pgnum.Int4From(req.Port),
		HealthCheckPath: defaultString(req.HealthCheckPath, DefaultHealthCheckPath),
		InternalUrl:     pgtext.From(req.InternalURL),
		ExternalUrl:     pgtext.From(req.ExternalURL),
		ApiSpecUrl:      pgtext.From(req.APISpecURL),
		DashboardUrl:    pgtext.From(req.DashboardURL),
		RunbookUrl:      pgtext.From(req.RunbookURL),
		SloTarget:       pgnum.NumericFrom(req.SLOTarget),
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(svc), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	services, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]Response, len(services))
	for i, svc := range services {
		resp[i] = toResponse(svc)
	}
	return resp, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Response, error) {
	svc, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(svc), nil
}

func (s *Service) AddEnvironment(ctx context.Context, serviceID string, req AddEnvironmentRequest) (EnvironmentResponse, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return EnvironmentResponse{}, ErrInvalidID
	}

	if _, err := s.repo.GetEnvironment(ctx, sID, req.Environment); err == nil {
		return EnvironmentResponse{}, ErrEnvironmentTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return EnvironmentResponse{}, err
	}

	replicasMin := req.ReplicasMin
	if replicasMin <= 0 {
		replicasMin = 1
	}
	replicasMax := req.ReplicasMax
	if replicasMax <= 0 {
		replicasMax = 1
	}

	env, err := s.repo.AddEnvironment(ctx, store.AddServiceEnvironmentParams{
		ServiceID:       sID,
		Environment:     req.Environment,
		ClusterName:     pgtext.From(req.ClusterName),
		Namespace:       pgtext.From(req.Namespace),
		ReplicasMin:     replicasMin,
		ReplicasMax:     replicasMax,
		CpuRequest:      pgtext.From(req.CPURequest),
		MemoryRequest:   pgtext.From(req.MemoryRequest),
		CurrentImageTag: pgtext.From(req.CurrentImageTag),
		Url:             pgtext.From(req.URL),
	})
	if err != nil {
		return EnvironmentResponse{}, err
	}

	return toEnvironmentResponse(env), nil
}

func (s *Service) ListEnvironments(ctx context.Context, serviceID string) ([]EnvironmentResponse, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return nil, ErrInvalidID
	}

	envs, err := s.repo.ListEnvironments(ctx, sID)
	if err != nil {
		return nil, err
	}

	resp := make([]EnvironmentResponse, len(envs))
	for i, e := range envs {
		resp[i] = toEnvironmentResponse(e)
	}
	return resp, nil
}

func (s *Service) AddDependency(ctx context.Context, serviceID string, req AddDependencyRequest) (DependencyResponse, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return DependencyResponse{}, ErrInvalidID
	}
	dependsOnID, err := uuidutil.Parse(req.DependsOnServiceID)
	if err != nil {
		return DependencyResponse{}, ErrInvalidID
	}

	dep, err := s.repo.AddDependency(ctx, store.AddServiceDependencyParams{
		ServiceID:          sID,
		DependsOnServiceID: dependsOnID,
		DependencyType:     pgtext.From(defaultString(req.DependencyType, DefaultDependencyType)),
		IsCritical:         req.IsCritical,
	})
	if err != nil {
		return DependencyResponse{}, err
	}

	return toDependencyResponse(dep), nil
}

// ListDependencies returns the services that serviceID depends on.
func (s *Service) ListDependencies(ctx context.Context, serviceID string) ([]DependencyResponse, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return nil, ErrInvalidID
	}

	deps, err := s.repo.ListDependencies(ctx, sID)
	if err != nil {
		return nil, err
	}

	resp := make([]DependencyResponse, len(deps))
	for i, d := range deps {
		resp[i] = toDependencyResponse(d)
	}
	return resp, nil
}

// ListDependents returns the services that depend on serviceID — the impact list to show
// before deleting it.
func (s *Service) ListDependents(ctx context.Context, serviceID string) ([]DependencyResponse, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return nil, ErrInvalidID
	}

	deps, err := s.repo.ListDependents(ctx, sID)
	if err != nil {
		return nil, err
	}

	resp := make([]DependencyResponse, len(deps))
	for i, d := range deps {
		resp[i] = toDependencyResponse(d)
	}
	return resp, nil
}

func (s *Service) RemoveDependency(ctx context.Context, serviceID, dependsOnServiceID string) error {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return ErrInvalidID
	}
	dependsOnID, err := uuidutil.Parse(dependsOnServiceID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.RemoveDependency(ctx, sID, dependsOnID)
}

func (s *Service) AddTag(ctx context.Context, serviceID string, req AddTagRequest) error {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.AddTag(ctx, sID, req.Tag)
}

func (s *Service) RemoveTag(ctx context.Context, serviceID, tag string) error {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return ErrInvalidID
	}

	return s.repo.RemoveTag(ctx, sID, tag)
}

func (s *Service) ListTags(ctx context.Context, serviceID string) ([]string, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return nil, ErrInvalidID
	}

	tags, err := s.repo.ListTags(ctx, sID)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Tag
	}
	return names, nil
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

func toResponse(svc store.Service) Response {
	return Response{
		ID:              uuidutil.String(svc.ID),
		Name:            svc.Name,
		Slug:            svc.Slug,
		Description:     pgtext.To(svc.Description),
		ServiceTypeID:   svc.ServiceTypeID,
		LifecycleID:     pgnum.Int2To(svc.LifecycleID),
		TierID:          pgnum.Int2To(svc.TierID),
		ProjectID:       uuidutil.String(svc.ProjectID),
		OwnerTeamID:     uuidutil.String(svc.OwnerTeamID),
		TechLeadID:      uuidutil.String(svc.TechLeadID),
		RepoURL:         svc.RepoUrl,
		RepoProviderID:  pgnum.Int2To(svc.RepoProviderID),
		DefaultBranch:   svc.DefaultBranch,
		LanguageID:      pgnum.Int2To(svc.LanguageID),
		Framework:       pgtext.To(svc.Framework),
		DockerfilePath:  svc.DockerfilePath,
		RegistryImage:   pgtext.To(svc.RegistryImage),
		CIPipelineURL:   pgtext.To(svc.CiPipelineUrl),
		GitopsRepoPath:  pgtext.To(svc.GitopsRepoPath),
		K8sResourceKind: svc.K8sResourceKind,
		Port:            pgnum.Int4To(svc.Port),
		HealthCheckPath: svc.HealthCheckPath,
		InternalURL:     pgtext.To(svc.InternalUrl),
		ExternalURL:     pgtext.To(svc.ExternalUrl),
		APISpecURL:      pgtext.To(svc.ApiSpecUrl),
		DashboardURL:    pgtext.To(svc.DashboardUrl),
		RunbookURL:      pgtext.To(svc.RunbookUrl),
		SLOTarget:       pgnum.NumericTo(svc.SloTarget),
		IsActive:        svc.IsActive,
	}
}

func toEnvironmentResponse(e store.ServiceEnvironment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:              uuidutil.String(e.ID),
		ServiceID:       uuidutil.String(e.ServiceID),
		Environment:     e.Environment,
		ClusterName:     pgtext.To(e.ClusterName),
		Namespace:       pgtext.To(e.Namespace),
		ReplicasMin:     e.ReplicasMin,
		ReplicasMax:     e.ReplicasMax,
		CPURequest:      pgtext.To(e.CpuRequest),
		MemoryRequest:   pgtext.To(e.MemoryRequest),
		CurrentImageTag: pgtext.To(e.CurrentImageTag),
		URL:             pgtext.To(e.Url),
	}
}

func toDependencyResponse(d store.ServiceDependency) DependencyResponse {
	return DependencyResponse{
		ServiceID:          uuidutil.String(d.ServiceID),
		DependsOnServiceID: uuidutil.String(d.DependsOnServiceID),
		DependencyType:     pgtext.To(d.DependencyType),
		IsCritical:         d.IsCritical,
	}
}
