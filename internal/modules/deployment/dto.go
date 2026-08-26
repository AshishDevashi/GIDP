package deployment

// Valid deployment status values (append-only lifecycle; see status transitions in service.go).
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusSucceeded  = "succeeded"
	StatusFailed     = "failed"
	StatusRolledBack = "rolled_back"
)

// CreateDeploymentRequest is the payload for recording a new deployment event.
type CreateDeploymentRequest struct {
	ServiceID   string `json:"service_id" binding:"required,uuid"`
	Environment string `json:"environment" binding:"required,max=50"`

	ImageTag         string `json:"image_tag" binding:"required,max=255"`
	PreviousImageTag string `json:"previous_image_tag"`
	GitCommitSHA     string `json:"git_commit_sha"`
	GitBranch        string `json:"git_branch"`

	TriggeredByUserID string `json:"triggered_by_user_id" binding:"omitempty,uuid"`
	TriggerType       string `json:"trigger_type" binding:"required,oneof=manual ci_auto rollback scheduled"`
	CIRunURL          string `json:"ci_run_url"`

	DeployStrategy  string `json:"deploy_strategy"`
	GitopsCommitSHA string `json:"gitops_commit_sha"`

	IsRollback                 bool   `json:"is_rollback"`
	RolledBackFromDeploymentID string `json:"rolled_back_from_deployment_id" binding:"omitempty,uuid"`
}

// UpdateStatusRequest is the payload for transitioning a deployment's status.
// This is the only mutation ever allowed on a deployment row after creation.
type UpdateStatusRequest struct {
	Status        string `json:"status" binding:"required,oneof=pending in_progress succeeded failed rolled_back"`
	FailureReason string `json:"failure_reason"`
}

// Response is the public representation of a deployment.
type Response struct {
	ID                         string `json:"id"`
	ServiceID                  string `json:"service_id"`
	Environment                string `json:"environment"`
	ImageTag                   string `json:"image_tag"`
	PreviousImageTag           string `json:"previous_image_tag,omitempty"`
	GitCommitSHA               string `json:"git_commit_sha,omitempty"`
	GitBranch                  string `json:"git_branch,omitempty"`
	TriggeredByUserID          string `json:"triggered_by_user_id,omitempty"`
	TriggerType                string `json:"trigger_type"`
	CIRunURL                   string `json:"ci_run_url,omitempty"`
	DeployStrategy             string `json:"deploy_strategy"`
	GitopsCommitSHA            string `json:"gitops_commit_sha,omitempty"`
	Status                     string `json:"status"`
	StartedAt                  string `json:"started_at,omitempty"`
	CompletedAt                string `json:"completed_at,omitempty"`
	FailureReason              string `json:"failure_reason,omitempty"`
	IsRollback                 bool   `json:"is_rollback"`
	RolledBackFromDeploymentID string `json:"rolled_back_from_deployment_id,omitempty"`
	CreatedAt                  string `json:"created_at"`
}
