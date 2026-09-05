package databases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	MinSizeMB             int32 = 10
	MaxSizeMB             int32 = 2000
	MaxInstanceCapacityMB int64 = 20000 // 20GB in MB

	statusActive   = "active"
	statusDeleting = "deleting"
	statusDeleted  = "deleted"
	statusFailed   = "failed"
)

var (
	ErrNotFound           = errors.New("managed database not found")
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidName        = errors.New("invalid database name: must be 1-63 characters, start with a letter or underscore, and contain only alphanumeric characters and underscores")
	ErrInvalidSize        = errors.New("invalid database size: must be between 10MB and 2000MB")
	ErrQuotaExceeded      = errors.New("insufficient capacity on db instance: requested size exceeds available quota")
	ErrAlreadyExists      = errors.New("database with this name already exists on the instance")
	ErrNoActiveDBInstance = errors.New("no active db instance found to host database")
	ErrInstanceNotReady   = errors.New("db instance is not ready or postgres container is not running")
)

var dbNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

type Service struct {
	repo           *Repository
	client         PostgresClient
	secretResolver *SecretResolver
	log            *slog.Logger
}

func NewService(repo *Repository, client PostgresClient, resolver *SecretResolver, log *slog.Logger) *Service {
	if client == nil {
		client = NewPostgresClient()
	}
	if resolver == nil {
		resolver = NewSecretResolver()
	}
	return &Service{
		repo:           repo,
		client:         client,
		secretResolver: resolver,
		log:            log,
	}
}

// CreateDatabase validates size and quota, provisions the role and database on PostgreSQL,
// and saves the metadata record.
func (s *Service) CreateDatabase(ctx context.Context, req CreateRequest, userID string) (Response, error) {
	// 1. Validate name
	req.Name = strings.TrimSpace(req.Name)
	if !dbNamePattern.MatchString(req.Name) {
		return Response{}, ErrInvalidName
	}

	// 2. Validate size
	if req.SizeMB < MinSizeMB || req.SizeMB > MaxSizeMB {
		return Response{}, ErrInvalidSize
	}

	// 3. Normalize username
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		req.Username = req.Name
	}
	if !dbNamePattern.MatchString(req.Username) {
		return Response{}, ErrInvalidName
	}

	// 4. Resolve DB instance
	var instance store.DbInstance
	var err error
	if req.DBInstanceID != "" {
		instUUID, err := uuidutil.Parse(req.DBInstanceID)
		if err != nil {
			return Response{}, ErrInvalidID
		}
		instance, err = s.repo.GetDBInstanceByID(ctx, instUUID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Response{}, ErrNoActiveDBInstance
			}
			return Response{}, err
		}
	} else {
		instances, err := s.repo.ListActiveDBInstances(ctx)
		if err != nil {
			return Response{}, err
		}
		if len(instances) == 0 {
			return Response{}, ErrNoActiveDBInstance
		}
		instance = instances[0]
	}

	if instance.Status != "running" || instance.ContainerStatus != "running" {
		return Response{}, ErrInstanceNotReady
	}

	// 5. Check quota on instance
	totalAllocated, err := s.repo.GetTotalAllocatedMBByInstanceID(ctx, instance.ID)
	if err != nil {
		return Response{}, err
	}

	capacityMB := MaxInstanceCapacityMB
	if instance.StorageGb > 0 {
		capacityMB = int64(instance.StorageGb) * 1000
	}

	if totalAllocated+int64(req.SizeMB) > capacityMB {
		return Response{}, ErrQuotaExceeded
	}

	// 6. Check uniqueness on instance
	existing, err := s.repo.GetByName(ctx, instance.ID, req.Name)
	if err == nil && existing.Status != statusDeleted {
		return Response{}, ErrAlreadyExists
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Response{}, err
	}

	// 7. Resolve admin connection info
	host := instance.PublicIp.String
	if host == "" {
		host = instance.PrivateIp.String
	}
	if host == "" {
		return Response{}, fmt.Errorf("db instance has no reachable IP address")
	}

	adminPassword, err := s.secretResolver.Resolve(ctx, instance.Region, instance.AdminSecretName)
	if err != nil {
		s.log.Error("failed to resolve admin password", "secret", instance.AdminSecretName, "error", err)
		return Response{}, fmt.Errorf("resolve admin credentials: %w", err)
	}

	adminCfg := AdminConnConfig{
		Host:     host,
		Port:     instance.PostgresPort,
		Username: instance.AdminUsername,
		Password: adminPassword,
		SSLMode:  "disable",
	}

	// 8. Provision database & role on PostgreSQL container
	if err := s.client.CreateDatabase(ctx, adminCfg, req.Name, req.Username, req.Password); err != nil {
		s.log.Error("failed to provision remote postgres database", "name", req.Name, "error", err)
		return Response{}, fmt.Errorf("provision postgres database: %w", err)
	}

	// 9. Construct user connection string
	connString := BuildConnectionString(host, instance.PostgresPort, req.Name, req.Username, req.Password)

	// 10. Persist metadata record
	var createdBy pgtype.UUID
	if userID != "" {
		createdBy, _ = uuidutil.Parse(userID)
	}

	created, err := s.repo.Create(ctx, store.CreateManagedDatabaseParams{
		DbInstanceID:     instance.ID,
		Name:             req.Name,
		Username:         req.Username,
		Password:         req.Password,
		AllocatedMb:      req.SizeMB,
		Status:           statusActive,
		ConnectionString: connString,
		CreatedBy:        createdBy,
	})
	if err != nil {
		// Attempt rollback drop on DB if persistence fails
		_ = s.client.DropDatabase(ctx, adminCfg, req.Name, req.Username)
		if isUniqueViolation(err) {
			return Response{}, ErrAlreadyExists
		}
		return Response{}, err
	}

	return toResponse(created), nil
}

// GetByID retrieves a managed database by ID.
func (s *Service) GetByID(ctx context.Context, id string) (Response, error) {
	dbUUID, err := uuidutil.Parse(id)
	if err != nil {
		return Response{}, ErrInvalidID
	}
	persisted, err := s.repo.GetByID(ctx, dbUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Response{}, ErrNotFound
		}
		return Response{}, err
	}
	return toResponse(persisted), nil
}

// List returns all active managed databases.
func (s *Service) List(ctx context.Context) ([]Response, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(list))
	for i, db := range list {
		responses[i] = toResponse(db)
	}
	return responses, nil
}

// ListByInstanceID returns all databases for a specific instance.
func (s *Service) ListByInstanceID(ctx context.Context, instanceID string) ([]Response, error) {
	instUUID, err := uuidutil.Parse(instanceID)
	if err != nil {
		return nil, ErrInvalidID
	}
	list, err := s.repo.ListByInstanceID(ctx, instUUID)
	if err != nil {
		return nil, err
	}
	responses := make([]Response, len(list))
	for i, db := range list {
		responses[i] = toResponse(db)
	}
	return responses, nil
}

// GetConnectionString returns connection parameters and ready-to-use DSN for a database.
func (s *Service) GetConnectionString(ctx context.Context, id string) (ConnectionStringResponse, error) {
	dbUUID, err := uuidutil.Parse(id)
	if err != nil {
		return ConnectionStringResponse{}, ErrInvalidID
	}
	db, err := s.repo.GetByID(ctx, dbUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ConnectionStringResponse{}, ErrNotFound
		}
		return ConnectionStringResponse{}, err
	}

	instance, err := s.repo.GetDBInstanceByID(ctx, db.DbInstanceID)
	if err != nil {
		return ConnectionStringResponse{}, fmt.Errorf("retrieve parent db instance: %w", err)
	}

	host := instance.PublicIp.String
	if host == "" {
		host = instance.PrivateIp.String
	}

	connString := db.ConnectionString
	if connString == "" {
		connString = BuildConnectionString(host, instance.PostgresPort, db.Name, db.Username, db.Password)
	}

	return ConnectionStringResponse{
		ConnectionString: connString,
		Host:             host,
		Port:             instance.PostgresPort,
		Database:         db.Name,
		Username:         db.Username,
		Password:         db.Password,
	}, nil
}

// DeleteDatabase drops the remote PostgreSQL database and role and marks the metadata record deleted.
func (s *Service) DeleteDatabase(ctx context.Context, id string) error {
	dbUUID, err := uuidutil.Parse(id)
	if err != nil {
		return ErrInvalidID
	}

	db, err := s.repo.GetByID(ctx, dbUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	instance, err := s.repo.GetDBInstanceByID(ctx, db.DbInstanceID)
	if err == nil && instance.Status == "running" && instance.ContainerStatus == "running" {
		host := instance.PublicIp.String
		if host == "" {
			host = instance.PrivateIp.String
		}
		if host != "" {
			adminPassword, err := s.secretResolver.Resolve(ctx, instance.Region, instance.AdminSecretName)
			if err == nil {
				adminCfg := AdminConnConfig{
					Host:     host,
					Port:     instance.PostgresPort,
					Username: instance.AdminUsername,
					Password: adminPassword,
					SSLMode:  "disable",
				}
				if dropErr := s.client.DropDatabase(ctx, adminCfg, db.Name, db.Username); dropErr != nil {
					s.log.Warn("failed to drop remote database, proceeding with metadata deletion", "db", db.Name, "error", dropErr)
				}
			}
		}
	}

	// Soft delete record in metadata table, freeing up quota
	if _, err := s.repo.SoftDelete(ctx, db.ID); err != nil {
		return err
	}

	return nil
}

// GetQuota returns the total allocated vs available MB for the instance.
func (s *Service) GetQuota(ctx context.Context, instanceID string) (QuotaResponse, error) {
	var totalAllocated int64
	var err error
	var capacityMB int64 = MaxInstanceCapacityMB

	if instanceID != "" {
		instUUID, err := uuidutil.Parse(instanceID)
		if err != nil {
			return QuotaResponse{}, ErrInvalidID
		}
		totalAllocated, err = s.repo.GetTotalAllocatedMBByInstanceID(ctx, instUUID)
		if err != nil {
			return QuotaResponse{}, err
		}
		inst, err := s.repo.GetDBInstanceByID(ctx, instUUID)
		if err == nil && inst.StorageGb > 0 {
			capacityMB = int64(inst.StorageGb) * 1000
		}
	} else {
		totalAllocated, err = s.repo.GetTotalAllocatedMB(ctx)
		if err != nil {
			return QuotaResponse{}, err
		}
	}

	available := capacityMB - totalAllocated
	if available < 0 {
		available = 0
	}

	return QuotaResponse{
		TotalCapacityMB: capacityMB,
		AllocatedMB:     totalAllocated,
		AvailableMB:     available,
	}, nil
}

// TeardownInstanceDatabases cleans up all databases belonging to a dbinstance during cascade teardown.
func (s *Service) TeardownInstanceDatabases(ctx context.Context, instanceID pgtype.UUID, host string, port int32, adminUser, adminPassword string) error {
	activeDatabases, err := s.repo.ListActiveByInstanceID(ctx, instanceID)
	if err != nil {
		s.log.Error("failed to list active databases during cascade teardown", "instance_id", uuidutil.String(instanceID), "error", err)
	}

	if host != "" && adminPassword != "" && len(activeDatabases) > 0 {
		adminCfg := AdminConnConfig{
			Host:     host,
			Port:     port,
			Username: adminUser,
			Password: adminPassword,
			SSLMode:  "disable",
		}
		for _, db := range activeDatabases {
			if dropErr := s.client.DropDatabase(ctx, adminCfg, db.Name, db.Username); dropErr != nil {
				s.log.Warn("failed to drop remote database during cascade teardown", "db", db.Name, "error", dropErr)
			}
		}
	}

	_, err = s.repo.SoftDeleteByInstanceID(ctx, instanceID)
	return err
}

func BuildConnectionString(host string, port int32, dbname, username, password string) string {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, password),
		Host:     net.JoinHostPort(host, strconv.Itoa(int(port))),
		Path:     dbname,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func toResponse(persisted store.ManagedDatabase) Response {
	return Response{
		ID:               uuidutil.String(persisted.ID),
		DBInstanceID:     uuidutil.String(persisted.DbInstanceID),
		Name:             persisted.Name,
		Username:         persisted.Username,
		AllocatedMB:      persisted.AllocatedMb,
		Status:           persisted.Status,
		ConnectionString: persisted.ConnectionString,
		CreatedBy:        uuidutil.String(persisted.CreatedBy),
		CreatedAt:        persisted.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        persisted.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
