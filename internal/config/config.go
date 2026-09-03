package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from the environment.
type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
	GitHubToken string

	DockerHubUsername  string
	DockerHubToken     string
	DockerHubNamespace string

	// Infrastructure provisioning (DB instances).
	AWSRegion                     string
	DBInstanceType                string
	DBInstanceAMISSMParameter     string
	DBInstanceEngineVersion       string
	DBInstanceStorageGB           int32
	DBInstanceRootVolumeGB        int32
	DBInstancePostgresPort        int32
	DBInstancePostgresImage       string
	DBInstanceAdminUsername       string
	DBInstanceContainerName       string
	DBInstanceDataDeviceName      string
	DBInstanceDataMountPoint      string
	DBInstanceSSHIngressCIDR      string
	DBInstancePostgresIngressCIDR string
	DBInstanceProvisionTimeout    time.Duration
	DBInstanceReadinessTimeout    time.Duration
	TerraformBinPath              string
	TerraformModuleDir            string
	TerraformWorkDir              string
	SSHKeyDir                     string
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() *Config {
	return &Config{
		Env:         getEnv("APP_ENV", "development"),
		Port:        getEnv("APP_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/wolf_platform?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:      getEnvDuration("JWT_TTL", 24*time.Hour),
		GitHubToken: getEnv("GITHUB_TOKEN", ""),

		DockerHubUsername:  getEnv("DOCKERHUB_USERNAME", ""),
		DockerHubToken:     getEnv("DOCKERHUB_TOKEN", ""),
		DockerHubNamespace: getEnv("DOCKERHUB_NAMESPACE", ""),

		AWSRegion:                     getEnv("AWS_REGION", "us-east-1"),
		DBInstanceType:                getEnv("DB_INSTANCE_TYPE", "t3.micro"),
		DBInstanceAMISSMParameter:     getEnv("DB_INSTANCE_AMI_SSM_PARAMETER", "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"),
		DBInstanceEngineVersion:       getEnv("DB_INSTANCE_ENGINE_VERSION", "16"),
		DBInstanceStorageGB:           getEnvInt32("DB_INSTANCE_STORAGE_GB", 20),
		DBInstanceRootVolumeGB:        getEnvInt32("DB_INSTANCE_ROOT_VOLUME_GB", 10),
		DBInstancePostgresPort:        getEnvInt32("DB_INSTANCE_POSTGRES_PORT", 5432),
		DBInstancePostgresImage:       getEnv("DB_INSTANCE_POSTGRES_IMAGE", "postgres:16-alpine"),
		DBInstanceAdminUsername:       getEnv("DB_INSTANCE_ADMIN_USERNAME", "gidp"),
		DBInstanceContainerName:       getEnv("DB_INSTANCE_CONTAINER_NAME", "gidp-postgres"),
		DBInstanceDataDeviceName:      getEnv("DB_INSTANCE_DATA_DEVICE_NAME", "/dev/sdf"),
		DBInstanceDataMountPoint:      getEnv("DB_INSTANCE_DATA_MOUNT_POINT", "/var/lib/gidp/postgres"),
		DBInstanceSSHIngressCIDR:      getEnv("DB_INSTANCE_SSH_INGRESS_CIDR", "0.0.0.0/0"),
		DBInstancePostgresIngressCIDR: getEnv("DB_INSTANCE_POSTGRES_INGRESS_CIDR", "0.0.0.0/0"),
		DBInstanceProvisionTimeout:    getEnvDuration("DB_INSTANCE_PROVISION_TIMEOUT", 20*time.Minute),
		DBInstanceReadinessTimeout:    getEnvDuration("DB_INSTANCE_READINESS_TIMEOUT", 10*time.Minute),
		TerraformBinPath:              getEnv("TERRAFORM_BIN_PATH", ""),
		TerraformModuleDir:            getEnv("TERRAFORM_MODULE_DIR", "deploy/terraform/modules/ec2-instance"),
		TerraformWorkDir:              getEnv("TERRAFORM_WORK_DIR", "deploy/terraform/.workspaces"),
		SSHKeyDir:                     getEnv("SSH_KEY_DIR", "deploy/keys"),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvInt32(key string, fallback int32) int32 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(parsed)
		}
	}
	return fallback
}
