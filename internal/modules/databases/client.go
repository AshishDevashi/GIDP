package databases

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// AdminConnConfig holds the parameters needed to connect to PostgreSQL as an administrator.
type AdminConnConfig struct {
	Host     string
	Port     int32
	Username string
	Password string
	SSLMode  string
}

// DSN returns the connection string for the administrative database.
func (c AdminConnConfig) DSN(database string) string {
	if database == "" {
		database = "postgres"
	}
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.Username, c.Password),
		Host:     net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port))),
		Path:     database,
		RawQuery: fmt.Sprintf("sslmode=%s&connect_timeout=10", sslMode),
	}
	return u.String()
}

// PostgresClient handles remote PostgreSQL administrative operations.
type PostgresClient interface {
	CreateDatabase(ctx context.Context, cfg AdminConnConfig, dbName, username, password string) error
	DropDatabase(ctx context.Context, cfg AdminConnConfig, dbName, username string) error
}

type defaultPostgresClient struct{}

func NewPostgresClient() PostgresClient {
	return &defaultPostgresClient{}
}

// CreateDatabase provisions a new PostgreSQL role and database, granting full permissions.
func (c *defaultPostgresClient) CreateDatabase(ctx context.Context, cfg AdminConnConfig, dbName, username, password string) error {
	adminDSN := cfg.DSN("postgres")
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return fmt.Errorf("connect to admin postgres: %w", err)
	}
	defer conn.Close(context.Background())

	safeUser := pgx.Identifier{username}.Sanitize()
	safeDB := pgx.Identifier{dbName}.Sanitize()
	escapedPassword := strings.ReplaceAll(password, "'", "''")

	// 1. Create or update the role with login credentials.
	roleSQL := fmt.Sprintf(`
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = %s) THEN
    CREATE ROLE %s WITH LOGIN PASSWORD '%s';
  ELSE
    ALTER ROLE %s WITH LOGIN PASSWORD '%s';
  END IF;
END
$$;`, quoteLiteral(username), safeUser, escapedPassword, safeUser, escapedPassword)

	if _, err := conn.Exec(ctx, roleSQL); err != nil {
		return fmt.Errorf("create or alter role %s: %w", username, err)
	}

	// 2. Create the database owned by the role (must be run outside transaction).
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s OWNER %s;", safeDB, safeUser)
	if _, err := conn.Exec(ctx, createDBSQL); err != nil {
		// If database already exists, skip or return error.
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("create database %s: %w", dbName, err)
		}
	}

	// 3. Grant privileges on the database to the role.
	grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", safeDB, safeUser)
	if _, err := conn.Exec(ctx, grantSQL); err != nil {
		return fmt.Errorf("grant database privileges: %w", err)
	}

	// 4. Connect to the newly created database to configure schema-level permissions.
	newDBDSN := cfg.DSN(dbName)
	newDBConn, err := pgx.Connect(ctx, newDBDSN)
	if err != nil {
		// Even if connecting to the new DB for schema grants fails, the DB is created.
		// Try to continue, or return descriptive error.
		return fmt.Errorf("connect to new database for schema grants: %w", err)
	}
	defer newDBConn.Close(context.Background())

	schemaSQL := fmt.Sprintf(`
GRANT ALL ON SCHEMA public TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO %s;
`, safeUser, safeUser, safeUser, safeUser)

	if _, err := newDBConn.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("grant schema privileges: %w", err)
	}

	return nil
}

// DropDatabase terminates connections, drops the database, and removes the role.
func (c *defaultPostgresClient) DropDatabase(ctx context.Context, cfg AdminConnConfig, dbName, username string) error {
	adminDSN := cfg.DSN("postgres")
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return fmt.Errorf("connect to admin postgres: %w", err)
	}
	defer conn.Close(context.Background())

	safeDB := pgx.Identifier{dbName}.Sanitize()
	safeUser := pgx.Identifier{username}.Sanitize()

	// 1. Terminate all active backend connections to the target database.
	termSQL := "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid();"
	_, _ = conn.Exec(ctx, termSQL, dbName)

	// 2. Drop the database with FORCE if supported, otherwise standard drop.
	dropDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE);", safeDB)
	if _, err := conn.Exec(ctx, dropDBSQL); err != nil {
		fallbackDropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", safeDB)
		if _, fbErr := conn.Exec(ctx, fallbackDropSQL); fbErr != nil {
			return fmt.Errorf("drop database %s: %w", dbName, fbErr)
		}
	}

	// 3. Drop the role if specified and not a system role.
	if username != "" && username != "postgres" && username != cfg.Username {
		dropRoleSQL := fmt.Sprintf("DROP ROLE IF EXISTS %s;", safeUser)
		if _, err := conn.Exec(ctx, dropRoleSQL); err != nil {
			// Role drop may fail if referenced elsewhere; log/return error.
			return fmt.Errorf("drop role %s: %w", username, err)
		}
	}

	return nil
}

func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// SecretResolver resolves SSM secret parameters with in-memory caching.
type SecretResolver struct {
	cache sync.Map
}

func NewSecretResolver() *SecretResolver {
	return &SecretResolver{}
}

// Resolve fetches the secret value from SSM or cache, with environment fallback.
func (r *SecretResolver) Resolve(ctx context.Context, region, secretName string) (string, error) {
	if secretName == "" {
		if envPass := os.Getenv("DB_INSTANCE_ADMIN_PASSWORD"); envPass != "" {
			return envPass, nil
		}
		return "", fmt.Errorf("no secret name provided")
	}

	if val, ok := r.cache.Load(secretName); ok {
		if s, ok := val.(string); ok && s != "" {
			return s, nil
		}
	}

	// Try AWS CLI SSM get-parameter.
	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxTimeout, "aws", "ssm", "get-parameter",
		"--name", secretName,
		"--with-decryption",
		"--region", region,
		"--query", "Parameter.Value",
		"--output", "text",
	)
	out, err := cmd.Output()
	if err == nil {
		password := strings.TrimSpace(string(out))
		if password != "" {
			r.cache.Store(secretName, password)
			return password, nil
		}
	}

	// Fallback to environment variable if SSM resolution is unavailable (e.g. dev/test mode).
	if envPass := os.Getenv("DB_INSTANCE_ADMIN_PASSWORD"); envPass != "" {
		r.cache.Store(secretName, envPass)
		return envPass, nil
	}

	return "", fmt.Errorf("resolve secret %s: %w", secretName, err)
}

// Cache stores a known secret value in the resolver.
func (r *SecretResolver) Cache(secretName, value string) {
	if secretName != "" && value != "" {
		r.cache.Store(secretName, value)
	}
}
