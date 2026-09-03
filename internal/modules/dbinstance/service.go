package dbinstance

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/sshkey"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Instance lifecycle states persisted in db_instances.status.
const (
	statusProvisioning = "provisioning"
	statusFailed       = "failed"
	statusDeleting     = "deleting"

	// PostgreSQL container states persisted in db_instances.container_status.
	containerPending = "pending"
	containerRunning = "running"
	containerFailed  = "failed"

	namePrefix       = "gidp-db-"
	secretNameFormat = "/gidp/db-instances/%s/admin-password"

	readinessPollInterval = 10 * time.Second
)

var (
	ErrNotFound      = errors.New("db instance not found")
	ErrInvalidID     = errors.New("invalid id")
	ErrNotConfigured = errors.New("infrastructure provisioning is not configured")
	ErrBusy          = errors.New("db instance is already being provisioned or deleted")
	ErrAlreadyExists = errors.New("a db instance already exists; only one is allowed at a time")
)

// Config carries the static infrastructure settings used for every DB instance.
type Config struct {
	Region              string
	InstanceType        string
	AMISSMParameter     string
	Engine              string
	EngineVersion       string
	StorageGB           int32
	RootVolumeGB        int32
	PostgresPort        int32
	PostgresImage       string
	AdminUsername       string
	ContainerName       string
	DataDeviceName      string
	DataMountPoint      string
	SSHIngressCIDR      string
	PostgresIngressCIDR string
	ModuleDir           string
	WorkDir             string
	KeyDir              string
	TerraformBinPath    string
	ProvisionTimeout    time.Duration
	ReadinessTimeout    time.Duration
}

type Service struct {
	repo        *Repository
	provisioner *provisioner
	cfg         Config
	log         *slog.Logger
}

func NewService(repo *Repository, prov *provisioner, cfg Config, log *slog.Logger) *Service {
	return &Service{repo: repo, provisioner: prov, cfg: cfg, log: log}
}

// Create registers the single DB instance and provisions its EC2 host in the
// background. The returned instance is in the "provisioning" state; clients
// poll GetByID. Only one live instance is allowed at a time.
func (s *Service) Create(ctx context.Context, userID string) (Response, error) {
	if s.provisioner == nil {
		return Response{}, ErrNotConfigured
	}

	createdBy, err := uuidutil.Parse(userID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	active, err := s.repo.CountActive(ctx)
	if err != nil {
		return Response{}, err
	}
	if active > 0 {
		return Response{}, ErrAlreadyExists
	}

	name, err := generateName()
	if err != nil {
		return Response{}, err
	}

	adminPassword, err := generatePassword()
	if err != nil {
		return Response{}, err
	}
	adminSecretName := fmt.Sprintf(secretNameFormat, name)

	userData, err := renderBootstrap(bootstrapParams{
		Region:          s.cfg.Region,
		AdminSecretName: adminSecretName,
		AdminUsername:   s.cfg.AdminUsername,
		PostgresImage:   s.cfg.PostgresImage,
		PostgresPort:    s.cfg.PostgresPort,
		DataDeviceName:  s.cfg.DataDeviceName,
		DataMountPoint:  s.cfg.DataMountPoint,
		ContainerName:   s.cfg.ContainerName,
	})
	if err != nil {
		return Response{}, err
	}

	key, err := sshkey.Ensure(s.cfg.KeyDir, name)
	if err != nil {
		return Response{}, err
	}

	created, err := s.repo.Create(ctx, store.CreateDBInstanceParams{
		Name:            name,
		Description:     "PostgreSQL host provisioned by GIDP",
		Engine:          s.cfg.Engine,
		EngineVersion:   s.cfg.EngineVersion,
		Provider:        "aws",
		Region:          s.cfg.Region,
		InstanceType:    s.cfg.InstanceType,
		StorageGb:       s.cfg.StorageGB,
		Workspace:       name,
		SshKeyName:      key.Name,
		AdminUsername:   s.cfg.AdminUsername,
		AdminSecretName: adminSecretName,
		PostgresPort:    s.cfg.PostgresPort,
		PostgresImage:   s.cfg.PostgresImage,
		CreatedBy:       createdBy,
	})
	if err != nil {
		_ = sshkey.Remove(s.cfg.KeyDir, name)
		if isUniqueViolation(err) {
			return Response{}, ErrAlreadyExists
		}
		return Response{}, err
	}

	go s.provisionAsync(created, key.PublicKey, adminSecretName, adminPassword, userData)

	return toResponse(created), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Response, error) {
	instanceID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(persisted), nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	instances, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(instances))
	for i, persisted := range instances {
		responses[i] = toResponse(persisted)
	}
	return responses, nil
}

// Delete tears down the underlying AWS resources in the background and soft
// deletes the record once Terraform has destroyed the workspace.
func (s *Service) Delete(ctx context.Context, id string) error {
	if s.provisioner == nil {
		return ErrNotConfigured
	}

	instanceID, err := uuidutil.Parse(id)
	if err != nil {
		return ErrInvalidID
	}

	persisted, err := s.repo.GetByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if persisted.Status == statusProvisioning || persisted.Status == statusDeleting {
		return ErrBusy
	}

	if err := s.setStatus(ctx, persisted.ID, statusDeleting, ""); err != nil {
		return err
	}

	go s.destroyAsync(persisted)

	return nil
}

func (s *Service) provisionAsync(instance store.DbInstance, publicKey, adminSecretName, adminPassword, userData string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ProvisionTimeout)
	defer cancel()

	id := uuidutil.String(instance.ID)
	s.log.Info("provisioning db instance", "id", id, "name", instance.Name)

	result, err := s.provisioner.provision(ctx, provisionInput{
		name:            instance.Name,
		workspace:       instance.Workspace,
		keyName:         instance.SshKeyName,
		publicKey:       publicKey,
		storageGB:       instance.StorageGb,
		adminSecretName: adminSecretName,
		adminPassword:   adminPassword,
		userData:        userData,
	})
	if err != nil {
		s.log.Error("db instance provisioning failed", "id", id, "error", err)
		if statusErr := s.setStatus(ctx, instance.ID, statusFailed, err.Error()); statusErr != nil {
			s.log.Error("failed to record provisioning failure", "id", id, "error", statusErr)
		}
		return
	}

	if _, err := s.repo.MarkProvisioned(ctx, store.MarkDBInstanceProvisionedParams{
		ID:                 instance.ID,
		ProviderInstanceID: pgtext.From(result.instanceID),
		AvailabilityZone:   pgtext.From(result.availabilityZone),
		PublicIp:           pgtext.From(result.publicIP),
		PrivateIp:          pgtext.From(result.privateIP),
		SecurityGroupID:    pgtext.From(result.securityGroupID),
		VolumeID:           pgtext.From(result.volumeID),
	}); err != nil {
		s.log.Error("failed to persist provisioned db instance", "id", id, "error", err)
		return
	}

	s.log.Info("db instance provisioned", "id", id, "ec2_instance_id", result.instanceID)

	s.awaitPostgres(instance, result.publicIP)
}

// awaitPostgres polls the container port until cloud-init has finished
// installing Docker and PostgreSQL accepts connections. terraform apply returns
// as soon as the instance boots, so this is the only honest readiness signal.
func (s *Service) awaitPostgres(instance store.DbInstance, host string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ReadinessTimeout)
	defer cancel()

	id := uuidutil.String(instance.ID)
	if host == "" {
		s.setContainerStatus(ctx, instance.ID, containerFailed, "instance has no reachable address")
		return
	}

	address := net.JoinHostPort(host, strconv.Itoa(int(instance.PostgresPort)))
	dialer := net.Dialer{Timeout: 5 * time.Second}

	for {
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = conn.Close()
			s.setContainerStatus(ctx, instance.ID, containerRunning, "")
			s.log.Info("postgres container is accepting connections", "id", id, "address", address)
			return
		}

		select {
		case <-ctx.Done():
			s.log.Error("postgres container never became reachable", "id", id, "address", address)
			s.setContainerStatus(context.Background(), instance.ID, containerFailed,
				"postgres did not accept connections before the readiness timeout")
			return
		case <-time.After(readinessPollInterval):
		}
	}
}

func (s *Service) setContainerStatus(ctx context.Context, id pgtype.UUID, status, message string) {
	if _, err := s.repo.MarkContainerStatus(ctx, store.MarkDBInstanceContainerStatusParams{
		ID:              id,
		ContainerStatus: status,
		StatusMessage:   pgtext.From(message),
	}); err != nil {
		s.log.Error("failed to record container status", "id", uuidutil.String(id), "error", err)
	}
}

func (s *Service) destroyAsync(instance store.DbInstance) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ProvisionTimeout)
	defer cancel()

	id := uuidutil.String(instance.ID)

	if err := s.provisioner.destroy(ctx, instance.Workspace); err != nil {
		s.log.Error("db instance teardown failed", "id", id, "error", err)
		if statusErr := s.setStatus(ctx, instance.ID, statusFailed, err.Error()); statusErr != nil {
			s.log.Error("failed to record teardown failure", "id", id, "error", statusErr)
		}
		return
	}

	if _, err := s.repo.SoftDelete(ctx, instance.ID); err != nil {
		s.log.Error("failed to soft delete db instance", "id", id, "error", err)
		return
	}
	if err := sshkey.Remove(s.cfg.KeyDir, instance.SshKeyName); err != nil {
		s.log.Warn("failed to remove ssh key pair", "id", id, "error", err)
	}

	s.log.Info("db instance destroyed", "id", id)
}

func (s *Service) setStatus(ctx context.Context, id pgtype.UUID, status, message string) error {
	rows, err := s.repo.MarkStatus(ctx, store.MarkDBInstanceStatusParams{
		ID:            id,
		Status:        status,
		StatusMessage: pgtext.From(message),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func generateName() (string, error) {
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("dbinstance: generate name: %w", err)
	}
	return namePrefix + hex.EncodeToString(buf), nil
}

// generatePassword returns a URL-safe secret, avoiding characters that would
// need escaping in the container env file or a connection string.
func generatePassword() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("dbinstance: generate password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toResponse(persisted store.DbInstance) Response {
	return Response{
		ID:                 uuidutil.String(persisted.ID),
		Name:               persisted.Name,
		Description:        persisted.Description,
		Engine:             persisted.Engine,
		EngineVersion:      persisted.EngineVersion,
		Provider:           persisted.Provider,
		Region:             persisted.Region,
		InstanceType:       persisted.InstanceType,
		StorageGB:          persisted.StorageGb,
		Status:             persisted.Status,
		ContainerStatus:    persisted.ContainerStatus,
		StatusMessage:      pgtext.To(persisted.StatusMessage),
		Workspace:          persisted.Workspace,
		SSHKeyName:         persisted.SshKeyName,
		AdminUsername:      persisted.AdminUsername,
		AdminSecretName:    persisted.AdminSecretName,
		PostgresPort:       persisted.PostgresPort,
		PostgresImage:      persisted.PostgresImage,
		ProviderInstanceID: pgtext.To(persisted.ProviderInstanceID),
		AvailabilityZone:   pgtext.To(persisted.AvailabilityZone),
		PublicIP:           pgtext.To(persisted.PublicIp),
		PrivateIP:          pgtext.To(persisted.PrivateIp),
		SecurityGroupID:    pgtext.To(persisted.SecurityGroupID),
		VolumeID:           pgtext.To(persisted.VolumeID),
		CreatedBy:          uuidutil.String(persisted.CreatedBy),
		CreatedAt:          persisted.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:          persisted.UpdatedAt.Time.Format(time.RFC3339),
	}
}
