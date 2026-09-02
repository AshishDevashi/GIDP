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

	LifecycleID int16 `json:"lifecycle_id"`
	TierID      int16 `json:"tier_id"`

	DocsURL      string `json:"docs_url"`
	DashboardURL string `json:"dashboard_url"`
	RunbookURL   string `json:"runbook_url"`

	ParentProjectID string `json:"parent_project_id" binding:"omitempty,uuid"`
}

// UpdateProjectRequest is the payload for updating project metadata.
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required,max=150"`
	Slug        string `json:"slug" binding:"required,max=150"`
	Description string `json:"description"`

	ProjectType  string `json:"project_type"`
	Architecture string `json:"architecture"`

	OwnerTeamID string `json:"owner_team_id" binding:"required,uuid"`
	TechLeadID  string `json:"tech_lead_id" binding:"omitempty,uuid"`

	LifecycleID int16 `json:"lifecycle_id"`
	TierID      int16 `json:"tier_id"`

	DocsURL      string `json:"docs_url"`
	DashboardURL string `json:"dashboard_url"`
	RunbookURL   string `json:"runbook_url"`

	ParentProjectID string `json:"parent_project_id" binding:"omitempty,uuid"`
	IsActive        *bool  `json:"is_active"`
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

// AddDependencyRequest is the payload for declaring that a project depends on another project.
type AddDependencyRequest struct {
	DependsOnProjectID string `json:"depends_on_project_id" binding:"required,uuid"`
	DependencyType     string `json:"dependency_type"`
}

// DependencyResponse is the public representation of a project dependency edge.
type DependencyResponse struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"project_id"`
	DependsOnProjectID string `json:"depends_on_project_id"`
	DependencyType     string `json:"dependency_type"`
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
	LifecycleID     int16  `json:"lifecycle_id"`
	TierID          int16  `json:"tier_id,omitempty"`
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
