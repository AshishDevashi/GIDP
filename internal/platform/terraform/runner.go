// Package terraform wraps terraform-exec so application code can drive
// Terraform runs from Go, one isolated workspace directory per resource.
package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// workspaceNamePattern keeps generated workspace names safe to join onto the
// work root, preventing path traversal from caller supplied identifiers.
var workspaceNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

// RootConfig is a Terraform root module expressed in JSON (main.tf.json), so
// every input value is produced in Go instead of being declared in HCL.
type RootConfig struct {
	Terraform map[string]any `json:"terraform,omitempty"`
	Provider  map[string]any `json:"provider,omitempty"`
	Module    map[string]any `json:"module,omitempty"`
	Output    map[string]any `json:"output,omitempty"`
}

// Outputs holds the root module outputs of a completed run.
type Outputs map[string]tfexec.OutputMeta

// String returns the output value as a string, or "" when absent or non-string.
func (o Outputs) String(key string) string {
	meta, ok := o[key]
	if !ok {
		return ""
	}
	var value string
	if err := json.Unmarshal(meta.Value, &value); err != nil {
		return ""
	}
	return value
}

// Runner executes Terraform in per-workspace directories under workRoot.
type Runner struct {
	execPath string
	workRoot string
	logs     io.Writer
}

// NewRunner locates the Terraform binary and prepares the workspace root.
// execPath may be empty, in which case terraform is looked up on PATH.
func NewRunner(execPath, workRoot string, logs io.Writer) (*Runner, error) {
	if execPath == "" {
		found, err := exec.LookPath("terraform")
		if err != nil {
			return nil, fmt.Errorf("terraform: binary not found on PATH: %w", err)
		}
		execPath = found
	}

	absWorkRoot, err := filepath.Abs(workRoot)
	if err != nil {
		return nil, fmt.Errorf("terraform: resolve work root: %w", err)
	}
	if err := os.MkdirAll(absWorkRoot, 0o750); err != nil {
		return nil, fmt.Errorf("terraform: create work root: %w", err)
	}

	return &Runner{execPath: execPath, workRoot: absWorkRoot, logs: logs}, nil
}

// Apply writes cfg into the workspace directory and runs init + apply,
// returning the root module outputs.
func (r *Runner) Apply(ctx context.Context, workspace string, cfg RootConfig) (Outputs, error) {
	tf, err := r.prepare(workspace, &cfg)
	if err != nil {
		return nil, err
	}

	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		return nil, fmt.Errorf("terraform init: %w", err)
	}
	if err := tf.Apply(ctx); err != nil {
		return nil, fmt.Errorf("terraform apply: %w", err)
	}

	outputs, err := tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform output: %w", err)
	}
	return outputs, nil
}

// Destroy tears down everything tracked by the workspace state.
func (r *Runner) Destroy(ctx context.Context, workspace string) error {
	tf, err := r.prepare(workspace, nil)
	if err != nil {
		return err
	}

	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		return fmt.Errorf("terraform init: %w", err)
	}
	if err := tf.Destroy(ctx); err != nil {
		return fmt.Errorf("terraform destroy: %w", err)
	}
	return nil
}

// RemoveWorkspace deletes the workspace directory and its state.
func (r *Runner) RemoveWorkspace(workspace string) error {
	dir, err := r.workspaceDir(workspace)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

// WorkspaceDir returns the absolute directory Terraform runs in for workspace.
func (r *Runner) WorkspaceDir(workspace string) (string, error) {
	return r.workspaceDir(workspace)
}

func (r *Runner) prepare(workspace string, cfg *RootConfig) (*tfexec.Terraform, error) {
	dir, err := r.workspaceDir(workspace)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("terraform: create workspace dir: %w", err)
	}

	if cfg != nil {
		encoded, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("terraform: encode config: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.tf.json"), encoded, 0o600); err != nil {
			return nil, fmt.Errorf("terraform: write config: %w", err)
		}
	}

	tf, err := tfexec.NewTerraform(dir, r.execPath)
	if err != nil {
		return nil, fmt.Errorf("terraform: init executor: %w", err)
	}
	if r.logs != nil {
		tf.SetStdout(r.logs)
		tf.SetStderr(r.logs)
	}
	return tf, nil
}

func (r *Runner) workspaceDir(workspace string) (string, error) {
	if !workspaceNamePattern.MatchString(workspace) {
		return "", fmt.Errorf("terraform: invalid workspace name %q", workspace)
	}
	return filepath.Join(r.workRoot, workspace), nil
}
