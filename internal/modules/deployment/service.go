package deployment

import (
	"context"
	"errors"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const DefaultDeployStrategy = "rolling"

var (
	ErrNotFound  = errors.New("deployment not found")
	ErrInvalidID = errors.New("invalid id")
)

// Service contains the deployment module's business logic. Deployment rows are append-only:
// Create inserts a new event, UpdateStatus is the only mutation ever allowed afterwards.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateDeploymentRequest) (Response, error) {
	serviceID, err := uuidutil.Parse(req.ServiceID)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	triggeredByUserID, err := optionalUUID(req.TriggeredByUserID)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	rolledBackFromID, err := optionalUUID(req.RolledBackFromDeploymentID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	dep, err := s.repo.Create(ctx, store.CreateDeploymentParams{
		ServiceID:                  serviceID,
		Environment:                req.Environment,
		ImageTag:                   req.ImageTag,
		PreviousImageTag:           pgtext.From(req.PreviousImageTag),
		GitCommitSha:               pgtext.From(req.GitCommitSHA),
		GitBranch:                  pgtext.From(req.GitBranch),
		TriggeredByUserID:          triggeredByUserID,
		TriggerType:                req.TriggerType,
		CiRunUrl:                   pgtext.From(req.CIRunURL),
		DeployStrategy:             defaultString(req.DeployStrategy, DefaultDeployStrategy),
		GitopsCommitSha:            pgtext.From(req.GitopsCommitSHA),
		Status:                     StatusPending,
		StartedAt:                  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		IsRollback:                 req.IsRollback,
		RolledBackFromDeploymentID: rolledBackFromID,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(dep), nil
}

// UpdateStatus transitions a deployment's status. completed_at is stamped automatically
// when moving into a terminal state (succeeded/failed/rolled_back).
func (s *Service) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (Response, error) {
	depID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	var completedAt pgtype.Timestamptz
	if isTerminalStatus(req.Status) {
		completedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	dep, err := s.repo.UpdateStatus(ctx, store.UpdateDeploymentStatusParams{
		ID:            depID,
		Status:        req.Status,
		CompletedAt:   completedAt,
		FailureReason: pgtext.From(req.FailureReason),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}

	return toResponse(dep), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Response, error) {
	depID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	dep, err := s.repo.GetByID(ctx, depID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}

	return toResponse(dep), nil
}

// ListByService returns deploy history for a service, optionally filtered to one environment.
func (s *Service) ListByService(ctx context.Context, serviceID, environment string) ([]Response, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return nil, ErrInvalidID
	}

	var deployments []store.Deployment
	if environment != "" {
		deployments, err = s.repo.ListByServiceEnvironment(ctx, sID, environment)
	} else {
		deployments, err = s.repo.ListByService(ctx, sID)
	}
	if err != nil {
		return nil, err
	}

	resp := make([]Response, len(deployments))
	for i, d := range deployments {
		resp[i] = toResponse(d)
	}
	return resp, nil
}

// Current returns the latest successful deployment for a service+environment, i.e. what's
// actually running right now.
func (s *Service) Current(ctx context.Context, serviceID, environment string) (Response, error) {
	sID, err := uuidutil.Parse(serviceID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	dep, err := s.repo.GetCurrent(ctx, sID, environment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}

	return toResponse(dep), nil
}

func isTerminalStatus(status string) bool {
	switch status {
	case StatusSucceeded, StatusFailed, StatusRolledBack:
		return true
	default:
		return false
	}
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

func formatTimestamptz(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func toResponse(d store.Deployment) Response {
	return Response{
		ID:                         uuidutil.String(d.ID),
		ServiceID:                  uuidutil.String(d.ServiceID),
		Environment:                d.Environment,
		ImageTag:                   d.ImageTag,
		PreviousImageTag:           pgtext.To(d.PreviousImageTag),
		GitCommitSHA:               pgtext.To(d.GitCommitSha),
		GitBranch:                  pgtext.To(d.GitBranch),
		TriggeredByUserID:          uuidutil.String(d.TriggeredByUserID),
		TriggerType:                d.TriggerType,
		CIRunURL:                   pgtext.To(d.CiRunUrl),
		DeployStrategy:             d.DeployStrategy,
		GitopsCommitSHA:            pgtext.To(d.GitopsCommitSha),
		Status:                     d.Status,
		StartedAt:                  formatTimestamptz(d.StartedAt),
		CompletedAt:                formatTimestamptz(d.CompletedAt),
		FailureReason:              pgtext.To(d.FailureReason),
		IsRollback:                 d.IsRollback,
		RolledBackFromDeploymentID: uuidutil.String(d.RolledBackFromDeploymentID),
		CreatedAt:                  formatTimestamptz(d.CreatedAt),
	}
}
