package deployment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	platformkube "github.com/AshishDevashi/GIDP/internal/platform/kubernetes"
	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	statusPending   = "pending"
	statusDeploying = "deploying"
	statusRunning   = "running"
	statusFailed    = "failed"
	statusStopped   = "stopped"

	exposeTypeClusterIP    = "ClusterIP"
	exposeTypeNodePort     = "NodePort"
	exposeTypeLoadBalancer = "LoadBalancer"
	exposeTypeIngress      = "Ingress"

	defaultRolloutTimeout = 10 * time.Minute
	rolloutPollInterval   = 10 * time.Second
)

var (
	ErrInvalidID                  = errors.New("invalid id")
	ErrNotFound                   = errors.New("deployment not found")
	ErrRegistryNotFound           = errors.New("registry not found")
	ErrRepoNotFound               = errors.New("repository not found")
	ErrNoActiveDeploymentInstance = errors.New("an active deployment instance is required")
	ErrCapacityReached            = errors.New("deployment instance capacity has been reached")
	ErrNameTaken                  = errors.New("deployment name already exists in this namespace")
	ErrInvalidManifestInput       = errors.New("invalid deployment manifest input")
	ErrKubernetesApply            = errors.New("kubernetes apply failed")
)

type Service struct {
	repo *Repository
	log  *slog.Logger
}

func NewService(repo *Repository, log *slog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) Create(ctx context.Context, req CreateRequest, userID string) (Response, error) {
	createdBy, err := uuidutil.Parse(userID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	instance, err := s.repo.GetActiveDeploymentInstance(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNoActiveDeploymentInstance
		}
		return Response{}, err
	}
	if instance.Status != "active" || pgtext.To(instance.CredentialsRef) == "" {
		return Response{}, ErrNoActiveDeploymentInstance
	}

	count, err := s.repo.CountActiveByInstanceID(ctx, instance.ID)
	if err != nil {
		return Response{}, err
	}
	if count >= int64(instance.MaxDeployments) {
		return Response{}, ErrCapacityReached
	}

	registryID, err := uuidutil.Parse(req.RegistryID)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	registry, err := s.repo.GetRegistryByID(ctx, registryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrRegistryNotFound
		}
		return Response{}, err
	}
	if registry.Status != "active" {
		return Response{}, ErrRegistryNotFound
	}

	repoID := pgtype.UUID{}
	if strings.TrimSpace(req.RepoID) != "" {
		repoID, err = uuidutil.Parse(req.RepoID)
		if err != nil {
			return Response{}, ErrInvalidID
		}
		if _, err := s.repo.GetRepoByID(ctx, repoID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Response{}, ErrRepoNotFound
			}
			return Response{}, err
		}
	}

	normalized := normalizeRequest(req)
	imageRef, err := fullImageRef(registry, normalized.ImageName, normalized.ImageTag)
	if err != nil {
		return Response{}, err
	}
	rendered, err := RenderManifest(ManifestInput{
		Name:       normalized.Name,
		Namespace:  normalized.Namespace,
		Image:      imageRef,
		Replicas:   normalized.Replicas,
		Resources:  normalized.Resources,
		EnvVars:    normalized.EnvVars,
		SecretRefs: normalized.SecretRefs,
		Expose:     normalized.Expose,
	})
	if err != nil {
		return Response{}, err
	}

	resources, envVars, secretRefs, expose, err := requestJSON(normalized)
	if err != nil {
		return Response{}, err
	}

	created, err := s.repo.Create(ctx, store.CreateDeploymentParams{
		DeploymentInstanceID: instance.ID,
		RepoID:               repoID,
		RegistryID:           registry.ID,
		ImageName:            normalized.ImageName,
		ImageTag:             normalized.ImageTag,
		Name:                 normalized.Name,
		Namespace:            normalized.Namespace,
		Replicas:             normalized.Replicas,
		Resources:            resources,
		EnvVars:              envVars,
		SecretRefs:           secretRefs,
		Expose:               expose,
		K8sDeploymentName:    pgtext.From(rendered.DeploymentName),
		CreatedBy:            createdBy,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Response{}, ErrNameTaken
		}
		return Response{}, err
	}

	snapshot, err := json.Marshal(map[string]any{
		"manifest":        rendered.YAML,
		"deployment_name": rendered.DeploymentName,
		"service_name":    rendered.ServiceName,
		"ingress_name":    rendered.IngressName,
		"image":           imageRef,
	})
	if err != nil {
		return Response{}, err
	}
	if _, err := s.repo.CreateRevision(ctx, store.CreateDeploymentRevisionParams{
		DeploymentID:   created.ID,
		RevisionNo:     1,
		ImageTag:       normalized.ImageTag,
		ConfigSnapshot: snapshot,
		Status:         statusDeploying,
		TriggeredBy:    createdBy,
	}); err != nil {
		return Response{}, err
	}

	deploying, err := s.repo.MarkDeploying(ctx, created.ID)
	if err != nil {
		return Response{}, err
	}

	kubeClient, err := platformkube.NewClient(pgtext.To(instance.CredentialsRef))
	if err != nil {
		applyErr := fmt.Errorf("%w: %v", ErrKubernetesApply, err)
		s.fail(ctx, created.ID, applyErr)
		return Response{}, applyErr
	}
	if err := kubeClient.ApplyManifest(ctx, rendered.YAML); err != nil {
		applyErr := fmt.Errorf("%w: %v", ErrKubernetesApply, err)
		s.fail(ctx, created.ID, applyErr)
		return Response{}, applyErr
	}

	go s.watchRollout(deploying, pgtext.To(instance.CredentialsRef))

	response := toResponse(deploying)
	response.ImageRef = imageRef
	return response, nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(items))
	for i, item := range items {
		responses[i] = toResponse(item)
	}
	return responses, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Response, error) {
	deploymentID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, deploymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}

	response := toResponse(persisted)
	if revision, err := s.repo.LatestRevision(ctx, persisted.ID); err == nil {
		response.LatestRevision = toRevision(revision)
	}
	response.K8sStatus = s.liveK8sStatus(ctx, persisted)
	return response, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	deploymentID, err := uuidutil.Parse(id)
	if err != nil {
		return ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, deploymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	instance, err := s.repo.GetActiveDeploymentInstance(ctx)
	if err != nil || instance.Status != "active" || pgtext.To(instance.CredentialsRef) == "" {
		return ErrNoActiveDeploymentInstance
	}
	registry, err := s.repo.GetRegistryByID(ctx, persisted.RegistryID)
	if err != nil {
		return err
	}
	rendered, err := s.renderPersistedManifest(persisted, registry)
	if err != nil {
		s.fail(ctx, persisted.ID, err)
		return err
	}
	kubeClient, err := platformkube.NewClient(pgtext.To(instance.CredentialsRef))
	if err != nil {
		applyErr := fmt.Errorf("%w: %v", ErrKubernetesApply, err)
		s.fail(ctx, persisted.ID, applyErr)
		return applyErr
	}
	if err := kubeClient.DeleteManifest(ctx, rendered.YAML); err != nil {
		applyErr := fmt.Errorf("%w: %v", ErrKubernetesApply, err)
		s.fail(ctx, persisted.ID, applyErr)
		return applyErr
	}

	rows, err := s.repo.Delete(ctx, persisted.ID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) CountActiveByInstanceID(ctx context.Context, id pgtype.UUID) (int64, error) {
	return s.repo.CountActiveByInstanceID(ctx, id)
}

func (s *Service) watchRollout(deployment store.Deployment, credentialsRef string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRolloutTimeout)
	defer cancel()

	client, err := platformkube.NewClient(credentialsRef)
	if err != nil {
		s.fail(context.Background(), deployment.ID, err)
		return
	}
	name := pgtext.To(deployment.K8sDeploymentName)
	for {
		status, err := client.WorkloadStatus(ctx, deployment.Namespace, name)
		if err != nil {
			s.fail(context.Background(), deployment.ID, err)
			return
		}
		if status.State == statusRunning {
			if _, err := s.repo.MarkStatus(context.Background(), store.MarkDeploymentStatusParams{ID: deployment.ID, Status: statusRunning}); err != nil {
				s.log.Error("failed to mark deployment running", "id", uuidutil.String(deployment.ID), "error", err)
			}
			return
		}

		select {
		case <-ctx.Done():
			s.fail(context.Background(), deployment.ID, errors.New("deployment rollout timed out"))
			return
		case <-time.After(rolloutPollInterval):
		}
	}
}

func (s *Service) liveK8sStatus(ctx context.Context, persisted store.Deployment) *K8sStatus {
	instance, err := s.repo.GetActiveDeploymentInstance(ctx)
	if err != nil || pgtext.To(instance.CredentialsRef) == "" {
		return &K8sStatus{Error: ErrNoActiveDeploymentInstance.Error()}
	}
	client, err := platformkube.NewClient(pgtext.To(instance.CredentialsRef))
	if err != nil {
		return &K8sStatus{Error: err.Error()}
	}
	status, err := client.WorkloadStatus(ctx, persisted.Namespace, pgtext.To(persisted.K8sDeploymentName))
	if err != nil {
		return &K8sStatus{Error: err.Error()}
	}
	return &K8sStatus{
		State:               status.State,
		Replicas:            status.Replicas,
		ReadyReplicas:       status.ReadyReplicas,
		AvailableReplicas:   status.AvailableReplicas,
		UnavailableReplicas: status.UnavailableReplicas,
		Pods:                status.Pods,
		PodPhases:           status.PodPhases,
	}
}

func (s *Service) renderPersistedManifest(persisted store.Deployment, registry store.Registry) (RenderedManifest, error) {
	resources := ResourceRequest{}
	expose := ExposeRequest{}
	envVars := map[string]string{}
	secretRefs := map[string]SecretRef{}
	_ = json.Unmarshal(persisted.Resources, &resources)
	_ = json.Unmarshal(persisted.Expose, &expose)
	_ = json.Unmarshal(persisted.EnvVars, &envVars)
	_ = json.Unmarshal(persisted.SecretRefs, &secretRefs)
	imageRef, err := fullImageRef(registry, persisted.ImageName, persisted.ImageTag)
	if err != nil {
		return RenderedManifest{}, err
	}
	return RenderManifest(ManifestInput{
		Name:       persisted.Name,
		Namespace:  persisted.Namespace,
		Image:      imageRef,
		Replicas:   persisted.Replicas,
		Resources:  resources,
		EnvVars:    envVars,
		SecretRefs: secretRefs,
		Expose:     expose,
	})
}

func (s *Service) fail(ctx context.Context, id pgtype.UUID, cause error) {
	if _, err := s.repo.MarkStatus(ctx, store.MarkDeploymentStatusParams{ID: id, Status: statusFailed, LastError: pgtext.From(cause.Error())}); err != nil {
		s.log.Error("failed to mark deployment failed", "id", uuidutil.String(id), "error", err)
	}
}

func normalizeRequest(req CreateRequest) CreateRequest {
	req.ImageName = strings.TrimSpace(req.ImageName)
	req.ImageTag = strings.TrimSpace(req.ImageTag)
	req.Name = strings.TrimSpace(req.Name)
	req.Namespace = strings.TrimSpace(req.Namespace)
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Replicas <= 0 {
		req.Replicas = 1
	}
	if req.Resources.CPU == "" {
		req.Resources.CPU = "250m"
	}
	if req.Resources.Memory == "" {
		req.Resources.Memory = "256Mi"
	}
	if req.EnvVars == nil {
		req.EnvVars = map[string]string{}
	}
	if req.SecretRefs == nil {
		req.SecretRefs = map[string]SecretRef{}
	}
	if req.Expose.Type == "" {
		req.Expose.Type = exposeTypeClusterIP
	}
	if req.Expose.Port <= 0 {
		req.Expose.Port = 80
	}
	if req.Expose.TargetPort <= 0 {
		req.Expose.TargetPort = req.Expose.Port
	}
	if req.Expose.Path == "" {
		req.Expose.Path = "/"
	}
	return req
}

func requestJSON(req CreateRequest) ([]byte, []byte, []byte, []byte, error) {
	resources, err := json.Marshal(req.Resources)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	envVars, err := json.Marshal(req.EnvVars)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	secretRefs, err := json.Marshal(req.SecretRefs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	expose, err := json.Marshal(req.Expose)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return resources, envVars, secretRefs, expose, nil
}

func fullImageRef(registry store.Registry, imageName, imageTag string) (string, error) {
	registryURL := strings.TrimRight(pgtext.To(registry.RegistryUrl), "/")
	if registryURL == "" {
		return "", fmt.Errorf("%w: registry url is required", ErrInvalidManifestInput)
	}
	imageName = strings.Trim(strings.TrimSpace(imageName), "/")
	imageTag = strings.TrimSpace(imageTag)
	if imageName == "" || imageTag == "" {
		return "", fmt.Errorf("%w: image_name and image_tag are required", ErrInvalidManifestInput)
	}
	if imageName != strings.ToLower(imageName) {
		return "", fmt.Errorf("%w: image_name must be lowercase", ErrInvalidManifestInput)
	}
	return registryURL + "/" + imageName + ":" + imageTag, nil
}

func toResponse(persisted store.Deployment) Response {
	return Response{
		ID:                   uuidutil.String(persisted.ID),
		DeploymentInstanceID: uuidutil.String(persisted.DeploymentInstanceID),
		RepoID:               uuidutil.String(persisted.RepoID),
		RegistryID:           uuidutil.String(persisted.RegistryID),
		ImageName:            persisted.ImageName,
		ImageTag:             persisted.ImageTag,
		Name:                 persisted.Name,
		Namespace:            persisted.Namespace,
		Replicas:             persisted.Replicas,
		Resources:            rawJSON(persisted.Resources, `{"cpu":"250m","memory":"256Mi"}`),
		EnvVars:              rawJSON(persisted.EnvVars, `{}`),
		SecretRefs:           rawJSON(persisted.SecretRefs, `{}`),
		Expose:               rawJSON(persisted.Expose, `{"type":"ClusterIP","port":80}`),
		Status:               persisted.Status,
		CurrentRevision:      persisted.CurrentRevision,
		K8sDeploymentName:    pgtext.To(persisted.K8sDeploymentName),
		LastError:            pgtext.To(persisted.LastError),
		CreatedBy:            uuidutil.String(persisted.CreatedBy),
		CreatedAt:            persisted.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:            persisted.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func toRevision(persisted store.Deploymentrevision) *Revision {
	return &Revision{
		ID:        uuidutil.String(persisted.ID),
		Number:    persisted.RevisionNo,
		ImageTag:  persisted.ImageTag,
		Snapshot:  rawJSON(persisted.ConfigSnapshot, `{}`),
		Status:    persisted.Status,
		CreatedAt: persisted.CreatedAt.Time.Format(time.RFC3339),
	}
}

func rawJSON(data []byte, fallback string) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(data)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
