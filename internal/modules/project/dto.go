package project

// CreateProjectRequest is the payload for creating a new project.
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,max=150"`
	Slug        string `json:"slug" binding:"required,max=150"`
	Description string `json:"description"`

	ProjectType  string `json:"project_type"`
	Architecture string `json:"architecture"`

	OwnerTeamID string `json:"owner_team_id" binding:"required,uuid"`
	TechLeadID  string `json:"tech_lead_id" binding:"omitempty,uuid"`

	RepoURL       string `json:"repo_url"`
	RepoProvider  string `json:"repo_provider"`
	DefaultBranch string `json:"default_branch"`
	CIPipelineURL string `json:"ci_pipeline_url"`
	GitopsPath    string `json:"gitops_path"`

	Lifecycle string `json:"lifecycle"`
	Tier      string `json:"tier"`
	Language  string `json:"language"`
	Framework string `json:"framework"`

	DocsURL      string `json:"docs_url"`
	DashboardURL string `json:"dashboard_url"`
	RunbookURL   string `json:"runbook_url"`

	ParentProjectID string `json:"parent_project_id" binding:"omitempty,uuid"`
}

// AddEnvironmentRequest is the payload for registering a deployment environment on a project.
type AddEnvironmentRequest struct {
	Environment string `json:"environment" binding:"required,max=50"`
	ClusterName string `json:"cluster_name"`
	Namespace   string `json:"namespace"`
	URL         string `json:"url"`
	Replicas    int32  `json:"replicas"`
}

// LinkServiceRequest is the payload for attaching a service/component to a project.
type LinkServiceRequest struct {
	ServiceID string `json:"service_id" binding:"required,uuid"`
}

// Response is the public representation of a project.
type Response struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     string `json:"description,omitempty"`
	ProjectType     string `json:"project_type"`
	Architecture    string `json:"architecture,omitempty"`
	OwnerTeamID     string `json:"owner_team_id"`
	TechLeadID      string `json:"tech_lead_id,omitempty"`
	RepoURL         string `json:"repo_url,omitempty"`
	RepoProvider    string `json:"repo_provider,omitempty"`
	DefaultBranch   string `json:"default_branch"`
	CIPipelineURL   string `json:"ci_pipeline_url,omitempty"`
	GitopsPath      string `json:"gitops_path,omitempty"`
	Lifecycle       string `json:"lifecycle"`
	Tier            string `json:"tier,omitempty"`
	Language        string `json:"language,omitempty"`
	Framework       string `json:"framework,omitempty"`
	DocsURL         string `json:"docs_url,omitempty"`
	DashboardURL    string `json:"dashboard_url,omitempty"`
	RunbookURL      string `json:"runbook_url,omitempty"`
	ParentProjectID string `json:"parent_project_id,omitempty"`
	IsActive        bool   `json:"is_active"`
}

// EnvironmentResponse is the public representation of a project environment.
type EnvironmentResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Environment string `json:"environment"`
	ClusterName string `json:"cluster_name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	URL         string `json:"url,omitempty"`
	Replicas    int32  `json:"replicas"`
}
