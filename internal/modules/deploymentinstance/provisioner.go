package deploymentinstance

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AshishDevashi/GIDP/internal/platform/terraform"
)

const (
	terraformRequiredVersion = ">= 1.5.0"
	awsProviderSource        = "hashicorp/aws"
	awsProviderVersion       = "~> 5.60"
	terraformModuleKey       = "deployment_instance"
)

type provisionInput struct {
	name      string
	workspace string
	keyName   string
	publicKey string
	userData  string
}

type provisionResult struct {
	instanceID      string
	publicIP        string
	privateIP       string
	securityGroupID string
}

type provisioner struct {
	runner    *terraform.Runner
	moduleDir string
	cfg       Config
}

func newProvisioner(runner *terraform.Runner, cfg Config) (*provisioner, error) {
	moduleDir, err := filepath.Abs(cfg.ModuleDir)
	if err != nil {
		return nil, fmt.Errorf("deploymentinstance: resolve terraform module dir: %w", err)
	}
	return &provisioner{runner: runner, moduleDir: moduleDir, cfg: cfg}, nil
}

func (p *provisioner) provision(ctx context.Context, in provisionInput) (provisionResult, error) {
	moduleSource, err := p.moduleSource(in.workspace)
	if err != nil {
		return provisionResult{}, err
	}

	outputs, err := p.runner.Apply(ctx, in.workspace, p.rootConfig(in, moduleSource))
	if err != nil {
		return provisionResult{}, err
	}

	return provisionResult{
		instanceID:      outputs.String("instance_id"),
		publicIP:        outputs.String("public_ip"),
		privateIP:       outputs.String("private_ip"),
		securityGroupID: outputs.String("security_group_id"),
	}, nil
}

func (p *provisioner) destroy(ctx context.Context, workspace string) error {
	if err := p.runner.Destroy(ctx, workspace); err != nil {
		return err
	}
	return p.runner.RemoveWorkspace(workspace)
}

func (p *provisioner) moduleSource(workspace string) (string, error) {
	dir, err := p.runner.WorkspaceDir(workspace)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(dir, p.moduleDir)
	if err != nil {
		return "", fmt.Errorf("deploymentinstance: resolve module source: %w", err)
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, ".") {
		rel = "./" + rel
	}
	return rel, nil
}

func (p *provisioner) rootConfig(in provisionInput, moduleSource string) terraform.RootConfig {
	return terraform.RootConfig{
		Terraform: map[string]any{
			"required_version": terraformRequiredVersion,
			"required_providers": map[string]any{
				"aws": map[string]any{"source": awsProviderSource, "version": awsProviderVersion},
			},
		},
		Provider: map[string]any{
			"aws": map[string]any{"region": p.cfg.Region},
		},
		Module: map[string]any{
			terraformModuleKey: map[string]any{
				"source":                moduleSource,
				"name":                  in.name,
				"instance_type":         p.cfg.InstanceType,
				"ami_ssm_parameter":     p.cfg.AMISSMParameter,
				"key_name":              in.keyName,
				"public_key":            in.publicKey,
				"user_data":             in.userData,
				"root_volume_size_gb":   p.cfg.RootVolumeGB,
				"ssh_ingress_cidr":      p.cfg.SSHIngressCIDR,
				"kube_api_ingress_cidr": p.cfg.KubeAPIIngressCIDR,
				"http_ingress_cidr":     p.cfg.HTTPIngressCIDR,
				"https_ingress_cidr":    p.cfg.HTTPSIngressCIDR,
				"associate_public_ip":   true,
				"tags": map[string]string{
					"ManagedBy": "gidp",
					"Component": "deployment-instance",
					"Workspace": in.workspace,
				},
			},
		},
		Output: map[string]any{
			"instance_id":       moduleOutput("instance_id"),
			"public_ip":         moduleOutput("public_ip"),
			"private_ip":        moduleOutput("private_ip"),
			"security_group_id": moduleOutput("security_group_id"),
		},
	}
}

func moduleOutput(name string) map[string]any {
	return map[string]any{"value": fmt.Sprintf("${module.%s.%s}", terraformModuleKey, name)}
}
