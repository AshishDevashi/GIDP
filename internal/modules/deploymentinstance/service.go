package deploymentinstance

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/platform/sshkey"
	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/ssh"
)

const (
	statusProvisioning = "provisioning"
	statusActive       = "active"
	statusTerminated   = "terminated"
	statusFailed       = "failed"
	statusStopped      = "stopped"

	namePrefix            = "gidp-deploy-"
	kubeAPIPort           = "6443"
	readinessPollInterval = 10 * time.Second
)

var (
	ErrInvalidID         = errors.New("invalid id")
	ErrNotConfigured     = errors.New("infrastructure provisioning is not configured")
	ErrBusy              = errors.New("deployment instance is already being provisioned")
	ErrAlreadyExists     = errors.New("a deployment instance already exists; only one is allowed at a time")
	ErrActiveDeployments = errors.New("deployment instance has active deployments")
)

type Config struct {
	Region             string
	InstanceType       string
	AMISSMParameter    string
	RootVolumeGB       int32
	SSHIngressCIDR     string
	KubeAPIIngressCIDR string
	HTTPIngressCIDR    string
	HTTPSIngressCIDR   string
	ModuleDir          string
	WorkDir            string
	KeyDir             string
	CredentialsDir     string
	SSHUser            string
	TerraformBinPath   string
	ProvisionTimeout   time.Duration
	ReadinessTimeout   time.Duration
	MaxDeployments     int32
}

type Service struct {
	repo        *Repository
	provisioner *provisioner
	cfg         Config
	log         *slog.Logger
	deleteGuard func(ctx context.Context, instance store.Deploymentinstance) error
}

func NewService(repo *Repository, prov *provisioner, cfg Config, log *slog.Logger) *Service {
	return &Service{repo: repo, provisioner: prov, cfg: cfg, log: log}
}

func (s *Service) SetDeleteGuard(guard func(ctx context.Context, instance store.Deploymentinstance) error) {
	s.deleteGuard = guard
}

func (s *Service) Create(ctx context.Context, userID string) (Response, error) {
	if s.provisioner == nil {
		return Response{}, ErrNotConfigured
	}

	if existing, err := s.repo.GetLive(ctx); err == nil {
		if existing.Status == statusProvisioning {
			return toResponse(existing), nil
		}
		return Response{}, ErrAlreadyExists
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}

	createdBy, err := uuidutil.Parse(userID)
	if err != nil {
		return Response{}, ErrInvalidID
	}

	name, err := generateName()
	if err != nil {
		return Response{}, err
	}

	key, err := sshkey.Ensure(s.cfg.KeyDir, name)
	if err != nil {
		return Response{}, err
	}

	created, err := s.repo.Create(ctx, store.CreateDeploymentInstanceParams{
		Name:           name,
		Workspace:      name,
		SshKeyName:     key.Name,
		MaxDeployments: s.maxDeployments(),
		CreatedBy:      createdBy,
	})
	if err != nil {
		_ = sshkey.Remove(s.cfg.KeyDir, name)
		if isUniqueViolation(err) {
			if existing, getErr := s.repo.GetLive(ctx); getErr == nil && existing.Status == statusProvisioning {
				return toResponse(existing), nil
			}
			return Response{}, ErrAlreadyExists
		}
		return Response{}, err
	}

	go s.provisionAsync(created, key.PrivateKeyPath, key.PublicKey, renderBootstrap())

	return toResponse(created), nil
}

func (s *Service) Get(ctx context.Context) (*Response, error) {
	persisted, err := s.repo.GetLive(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	response := toResponse(persisted)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context) error {
	if s.provisioner == nil {
		return ErrNotConfigured
	}

	persisted, err := s.repo.GetLive(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if persisted.Status == statusProvisioning {
		return ErrBusy
	}
	if s.deleteGuard != nil {
		if err := s.deleteGuard(ctx, persisted); err != nil {
			return err
		}
	}

	go s.destroyAsync(persisted)
	return nil
}

func (s *Service) provisionAsync(instance store.Deploymentinstance, privateKeyPath, publicKey, userData string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ProvisionTimeout)
	defer cancel()

	id := uuidutil.String(instance.ID)
	s.log.Info("provisioning deployment instance", "id", id, "name", instance.Name)

	result, err := s.provisioner.provision(ctx, provisionInput{
		name:      instance.Name,
		workspace: instance.Workspace,
		keyName:   instance.SshKeyName,
		publicKey: publicKey,
		userData:  userData,
	})
	if err != nil {
		s.recordFailure(ctx, instance.ID, id, "deployment instance provisioning failed", err)
		return
	}

	apiServerURL := "https://" + net.JoinHostPort(result.publicIP, kubeAPIPort)
	if err := s.awaitK3s(result.publicIP); err != nil {
		s.recordFailure(context.Background(), instance.ID, id, "k3s readiness failed", err)
		return
	}

	credentialsRef, err := s.captureKubeconfig(context.Background(), result.publicIP, privateKeyPath, apiServerURL, instance.Name)
	if err != nil {
		s.recordFailure(context.Background(), instance.ID, id, "failed to capture kubeconfig", err)
		return
	}

	if _, err := s.repo.MarkProvisioned(context.Background(), store.MarkDeploymentInstanceProvisionedParams{
		ID:              instance.ID,
		Ec2InstanceID:   pgtext.From(result.instanceID),
		PublicIp:        pgtext.From(result.publicIP),
		PrivateIp:       pgtext.From(result.privateIP),
		ApiServerUrl:    pgtext.From(apiServerURL),
		CredentialsRef:  pgtext.From(credentialsRef),
		SecurityGroupID: pgtext.From(result.securityGroupID),
	}); err != nil {
		s.recordFailure(context.Background(), instance.ID, id, "failed to persist deployment instance", err)
		return
	}

	s.log.Info("deployment instance provisioned", "id", id, "ec2_instance_id", result.instanceID)
}

func (s *Service) awaitK3s(host string) error {
	if host == "" {
		return errors.New("instance has no reachable address")
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ReadinessTimeout)
	defer cancel()

	address := net.JoinHostPort(host, kubeAPIPort)
	dialer := net.Dialer{Timeout: 5 * time.Second}

	for {
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("k3s api did not accept connections before the readiness timeout: %w", ctx.Err())
		case <-time.After(readinessPollInterval):
		}
	}
}

func (s *Service) captureKubeconfig(ctx context.Context, host, privateKeyPath, apiServerURL, name string) (string, error) {
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("read ssh private key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("parse ssh private key: %w", err)
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(host, "22"), &ssh.ClientConfig{
		User:            s.sshUser(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	})
	if err != nil {
		return "", fmt.Errorf("dial ssh: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("open ssh session: %w", err)
	}
	defer session.Close()

	var stderr bytes.Buffer
	session.Stderr = &stderr
	contents, err := session.Output("sudo cat /etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return "", fmt.Errorf("read remote kubeconfig: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	rewritten := strings.ReplaceAll(string(contents), "https://127.0.0.1:6443", apiServerURL)
	rewritten = strings.ReplaceAll(rewritten, "https://localhost:6443", apiServerURL)

	if err := os.MkdirAll(s.cfg.CredentialsDir, 0o700); err != nil {
		return "", fmt.Errorf("create credentials dir: %w", err)
	}
	path := filepath.Join(s.cfg.CredentialsDir, name+".kubeconfig")
	if err := os.WriteFile(path, []byte(rewritten), 0o600); err != nil {
		return "", fmt.Errorf("write kubeconfig: %w", err)
	}
	return path, nil
}

func (s *Service) destroyAsync(instance store.Deploymentinstance) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ProvisionTimeout)
	defer cancel()

	id := uuidutil.String(instance.ID)
	if err := s.provisioner.destroy(ctx, instance.Workspace); err != nil {
		s.recordFailure(context.Background(), instance.ID, id, "deployment instance teardown failed", err)
		return
	}
	if ref := pgtext.To(instance.CredentialsRef); ref != "" {
		if err := os.Remove(ref); err != nil && !errors.Is(err, os.ErrNotExist) {
			s.log.Warn("failed to remove deployment kubeconfig", "id", id, "error", err)
		}
	}
	if _, err := s.repo.MarkTerminated(context.Background(), instance.ID); err != nil {
		s.log.Error("failed to mark deployment instance terminated", "id", id, "error", err)
		return
	}
	if err := sshkey.Remove(s.cfg.KeyDir, instance.SshKeyName); err != nil {
		s.log.Warn("failed to remove ssh key pair", "id", id, "error", err)
	}

	s.log.Info("deployment instance destroyed", "id", id)
}

func (s *Service) recordFailure(ctx context.Context, id pgtype.UUID, idText, message string, err error) {
	s.log.Error(message, "id", idText, "error", err)
	if statusErr := s.setStatus(ctx, id, statusFailed, err.Error()); statusErr != nil {
		s.log.Error("failed to record deployment instance failure", "id", idText, "error", statusErr)
	}
}

func (s *Service) setStatus(ctx context.Context, id pgtype.UUID, status, message string) error {
	rows, err := s.repo.MarkStatus(ctx, store.MarkDeploymentInstanceStatusParams{
		ID:        id,
		Status:    status,
		LastError: pgtext.From(message),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Service) sshUser() string {
	if s.cfg.SSHUser == "" {
		return "ec2-user"
	}
	return s.cfg.SSHUser
}

func (s *Service) maxDeployments() int32 {
	if s.cfg.MaxDeployments <= 0 {
		return 3
	}
	return s.cfg.MaxDeployments
}

func generateName() (string, error) {
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("deploymentinstance: generate name: %w", err)
	}
	return namePrefix + hex.EncodeToString(buf), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toResponse(persisted store.Deploymentinstance) Response {
	return Response{
		ID:              uuidutil.String(persisted.ID),
		Name:            persisted.Name,
		EC2InstanceID:   pgtext.To(persisted.Ec2InstanceID),
		PublicIP:        pgtext.To(persisted.PublicIp),
		PrivateIP:       pgtext.To(persisted.PrivateIp),
		APIServerURL:    pgtext.To(persisted.ApiServerUrl),
		AuthType:        persisted.AuthType,
		CredentialsRef:  pgtext.To(persisted.CredentialsRef),
		MaxDeployments:  persisted.MaxDeployments,
		Status:          persisted.Status,
		LastError:       pgtext.To(persisted.LastError),
		Workspace:       persisted.Workspace,
		SSHKeyName:      persisted.SshKeyName,
		SecurityGroupID: pgtext.To(persisted.SecurityGroupID),
		CreatedBy:       uuidutil.String(persisted.CreatedBy),
		CreatedAt:       persisted.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       persisted.UpdatedAt.Time.Format(time.RFC3339),
	}
}
