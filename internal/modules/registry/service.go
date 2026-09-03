package registry

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	dockerHubProviderID int16 = 1
	dockerHubRegistry         = "registry.hub.docker.com"
)

var (
	ErrNotConfigured       = errors.New("docker hub integration is not configured")
	ErrNamespaceRequired   = errors.New("namespace is required when no default docker hub namespace is configured")
	ErrUnauthorized        = errors.New("docker hub credentials are invalid or lack repository permissions")
	ErrRegistryTaken       = errors.New("a registry with this name already exists for this namespace")
	ErrRegistryInvalid     = errors.New("docker hub rejected the repository settings")
	ErrNotFound            = errors.New("registry not found")
	ErrInvalidID           = errors.New("invalid id")
	ErrDockerHubUnavailabe = errors.New("docker hub is currently unavailable")
)

type Service struct {
	repo             *Repository
	client           *dockerHubClient
	defaultNamespace string
}

func NewService(repo *Repository, client *dockerHubClient, defaultNamespace string) *Service {
	return &Service{repo: repo, client: client, defaultNamespace: defaultNamespace}
}

func (s *Service) Create(ctx context.Context, req CreateRequest, userID string) (Response, error) {
	if !s.client.configured() {
		return Response{}, ErrNotConfigured
	}

	createdBy, err := uuidutil.Parse(userID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	namespace, err := s.resolveNamespace(req.Namespace)
	if err != nil {
		return Response{}, err
	}

	if _, err := s.repo.GetByName(ctx, store.GetRegistryByNameParams{
		ProviderID: dockerHubProviderID,
		Namespace:  namespace,
		Name:       req.Name,
	}); err == nil {
		return Response{}, ErrRegistryTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}

	created, err := s.client.createRepository(ctx, namespace, req)
	if err != nil {
		return Response{}, mapDockerHubError(err)
	}

	visibility := "public"
	if created.IsPrivate {
		visibility = "private"
	}

	persisted, err := s.repo.Create(ctx, store.CreateRegistryParams{
		Name:        created.Name,
		Description: req.Description,
		ProviderID:  dockerHubProviderID,
		Namespace:   created.Namespace,
		RegistryUrl: pgtext.From(dockerHubRegistry),
		Visibility:  visibility,
		CreatedBy:   createdBy,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Response{}, ErrRegistryTaken
		}
		return Response{}, err
	}

	return toResponse(persisted), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Response, error) {
	registryID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(persisted), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	registries, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(registries))
	for i, persisted := range registries {
		responses[i] = toResponse(persisted)
	}
	return responses, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (Response, error) {
	registryID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	updated, err := s.repo.Update(ctx, store.UpdateRegistryParams{
		ID:          registryID,
		Description: req.Description,
		Visibility:  req.Visibility,
		Status:      req.Status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	registryID, err := uuidutil.Parse(id)
	if err != nil {
		return ErrInvalidID
	}

	persisted, err := s.repo.GetByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	if !s.client.configured() {
		return ErrNotConfigured
	}
	if err := s.client.deleteRepository(ctx, persisted.Namespace, persisted.Name); err != nil {
		return mapDockerHubError(err)
	}

	rows, err := s.repo.SoftDelete(ctx, registryID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) resolveNamespace(namespace string) (string, error) {
	namespace = strings.TrimSpace(namespace)
	if namespace != "" {
		return namespace, nil
	}
	if s.defaultNamespace != "" {
		return s.defaultNamespace, nil
	}
	return "", ErrNamespaceRequired
}

func mapDockerHubError(err error) error {
	var hubErr *dockerHubError
	if !errors.As(err, &hubErr) {
		return ErrDockerHubUnavailabe
	}
	switch hubErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusConflict:
		return ErrRegistryTaken
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		if strings.Contains(strings.ToLower(hubErr.Message), "already exists") {
			return ErrRegistryTaken
		}
		return ErrRegistryInvalid
	default:
		return ErrDockerHubUnavailabe
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toResponse(persisted store.Registry) Response {
	fullName := persisted.Namespace + "/" + persisted.Name
	return Response{
		ID:          uuidutil.String(persisted.ID),
		Name:        persisted.Name,
		Description: persisted.Description,
		ProviderID:  persisted.ProviderID,
		Namespace:   persisted.Namespace,
		FullName:    fullName,
		RegistryURL: pgtext.To(persisted.RegistryUrl),
		Visibility:  persisted.Visibility,
		Status:      persisted.Status,
		URL:         "https://hub.docker.com/r/" + fullName,
		PullCommand: "docker pull " + fullName,
		CreatedBy:   uuidutil.String(persisted.CreatedBy),
		CreatedAt:   persisted.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   persisted.UpdatedAt.Time.Format(time.RFC3339),
	}
}
