package repo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const githubProviderID int16 = 1

var (
	ErrNotConfigured        = errors.New("github integration is not configured")
	ErrUnauthorized         = errors.New("github credentials are invalid or lack repository permissions")
	ErrTemplateInaccessible = errors.New("template repository was not found, is not accessible, or is not marked as a GitHub template repository")
	ErrRepositoryInvalid    = errors.New("github repository already exists or its settings are invalid")
	ErrGitHubUnavailable    = errors.New("github is currently unavailable")
	ErrRepoTaken            = errors.New("repository already exists for this owner and provider")
	ErrNotFound             = errors.New("repository not found")
	ErrInvalidID            = errors.New("invalid id")
	ErrTemplateNotFound     = errors.New("no active template found for the requested language")
)

type Service struct {
	repo   *Repository
	client *githubClient
}

func NewService(repo *Repository, token string) *Service {
	return &Service{repo: repo, client: newGitHubClient(token)}
}

func (s *Service) Create(ctx context.Context, req CreateRequest, userID, username string) (Response, error) {
	createdBy, err := uuidutil.Parse(userID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	template, err := s.repo.GetTemplateBySlug(ctx, req.Language+"-app")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrTemplateNotFound
		}
		return Response{}, err
	}

	owner := username
	if req.Organization != "" {
		owner = req.Organization
	}
	visibility := "public"
	if req.Private {
		visibility = "private"
	}

	persisted, err := s.repo.Create(ctx, store.CreateRepoParams{
		Name:         req.Name,
		Owner:        owner,
		ProviderID:   githubProviderID,
		Visibility:   visibility,
		CreatedBy:    createdBy,
		TemplateUsed: pgtext.From(template.Slug),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Response{}, ErrRepoTaken
		}
		return Response{}, err
	}

	if _, err := s.repo.MarkCreating(ctx, persisted.ID); err != nil {
		return Response{}, err
	}

	if s.client.token == "" {
		return s.fail(ctx, persisted.ID, ErrNotConfigured.Error(), ErrNotConfigured)
	}

	githubRepo, err := s.client.createRepositoryFromTemplate(ctx, template.TemplateOwner, template.TemplateRepo, req)
	if err != nil {
		mappedErr := mapGitHubError(err)
		return s.fail(ctx, persisted.ID, githubFailureMessage(err, mappedErr), mappedErr)
	}

	defaultBranch := githubRepo.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	githubOwner := githubRepo.Owner.Login
	if githubOwner == "" {
		githubOwner = owner
	}
	active, err := s.repo.Activate(ctx, store.ActivateRepoParams{
		ID:            persisted.ID,
		FullName:      pgtext.From(githubRepo.FullName),
		Owner:         githubOwner,
		ExternalID:    pgtext.From(strconv.FormatInt(githubRepo.ID, 10)),
		Url:           pgtext.From(githubRepo.HTMLURL),
		CloneUrlSsh:   pgtext.From(githubRepo.SSHURL),
		CloneUrlHttps: pgtext.From(githubRepo.CloneURL),
		DefaultBranch: defaultBranch,
		Visibility:    visibility,
	})
	if err != nil {
		return Response{}, err
	}

	return toResponse(active), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Response, error) {
	repoID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, repoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(persisted), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	repos, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(repos))
	for i, persisted := range repos {
		responses[i] = toResponse(persisted)
	}
	return responses, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (Response, error) {
	repoID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	updated, err := s.repo.Update(ctx, store.UpdateRepoParams{
		ID:            repoID,
		Name:          req.Name,
		DefaultBranch: req.DefaultBranch,
		Visibility:    req.Visibility,
		TemplateUsed:  pgtext.From(req.TemplateUsed),
		Status:        req.Status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return Response{}, ErrRepoTaken
		}
		return Response{}, err
	}
	return toResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	repoID, err := uuidutil.Parse(id)
	if err != nil {
		return ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, repoID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if persisted.Status == "active" && persisted.ExternalID.Valid {
		if s.client.token == "" {
			return ErrNotConfigured
		}
		owner, name := persisted.Owner, persisted.Name
		if fullName := pgtext.To(persisted.FullName); fullName != "" {
			if fullNameOwner, fullNameRepo, ok := strings.Cut(fullName, "/"); ok {
				owner, name = fullNameOwner, fullNameRepo
			}
		}
		if err := s.client.deleteRepository(ctx, owner, name); err != nil {
			return mapGitHubDeleteError(err)
		}
	}

	rows, err := s.repo.Delete(ctx, repoID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func mapGitHubDeleteError(err error) error {
	var githubErr *githubError
	if !errors.As(err, &githubErr) {
		return ErrGitHubUnavailable
	}
	switch githubErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusConflict:
		return ErrRepositoryInvalid
	default:
		return ErrGitHubUnavailable
	}
}

func (s *Service) fail(ctx context.Context, id pgtype.UUID, message string, cause error) (Response, error) {
	failed, err := s.repo.Fail(ctx, store.FailRepoParams{ID: id, ErrorMessage: pgtext.From(message)})
	if err != nil {
		return Response{}, fmt.Errorf("mark repository failed: %w", err)
	}
	return toResponse(failed), cause
}

func mapGitHubError(err error) error {
	var githubErr *githubError
	if !errors.As(err, &githubErr) {
		return ErrGitHubUnavailable
	}
	switch githubErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrTemplateInaccessible
	case http.StatusUnprocessableEntity:
		return ErrRepositoryInvalid
	default:
		return ErrGitHubUnavailable
	}
}

func githubFailureMessage(err, fallback error) string {
	var githubErr *githubError
	if errors.As(err, &githubErr) && githubErr.Message != "" {
		return githubErr.Message
	}
	return fallback.Error()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toResponse(persisted store.Repo) Response {
	return Response{
		ID:            uuidutil.String(persisted.ID),
		Name:          persisted.Name,
		FullName:      pgtext.To(persisted.FullName),
		Owner:         persisted.Owner,
		ProviderID:    persisted.ProviderID,
		ExternalID:    pgtext.To(persisted.ExternalID),
		URL:           pgtext.To(persisted.Url),
		CloneURLSSH:   pgtext.To(persisted.CloneUrlSsh),
		CloneURLHTTPS: pgtext.To(persisted.CloneUrlHttps),
		DefaultBranch: persisted.DefaultBranch,
		Visibility:    persisted.Visibility,
		TemplateUsed:  pgtext.To(persisted.TemplateUsed),
		Status:        persisted.Status,
		ErrorMessage:  pgtext.To(persisted.ErrorMessage),
		CreatedBy:     uuidutil.String(persisted.CreatedBy),
		CreatedAt:     persisted.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:     persisted.UpdatedAt.Time.Format(time.RFC3339),
	}
}
