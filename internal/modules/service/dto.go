package service

// CreateServiceRequest is the payload for registering a new microservice/component.
type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required,max=150"`
	Slug        string `json:"slug" binding:"required,max=150"`
	Description string `json:"description"`

	ServiceTypeID int16 `json:"service_type_id"`
	LifecycleID   int16 `json:"lifecycle_id"`
	TierID        int16 `json:"tier_id"`

	ProjectID   string `json:"project_id" binding:"omitempty,uuid"`
	OwnerTeamID string `json:"owner_team_id" binding:"required,uuid"`
	TechLeadID  string `json:"tech_lead_id" binding:"omitempty,uuid"`

	RepoURL        string `json:"repo_url" binding:"required"`
	RepoProviderID int16  `json:"repo_provider_id"`
	DefaultBranch  string `json:"default_branch"`
	LanguageID     int16  `json:"language_id"`
	Framework      string `json:"framework"`

	DockerfilePath string `json:"dockerfile_path"`
	RegistryImage  string `json:"registry_image"`
	CIPipelineURL  string `json:"ci_pipeline_url"`

	GitopsRepoPath  string `json:"gitops_repo_path"`
	K8sResourceKind string `json:"k8s_resource_kind"`

	Port            int32  `json:"port"`
	HealthCheckPath string `json:"health_check_path"`
	InternalURL     string `json:"internal_url"`
	ExternalURL     string `json:"external_url"`

	APISpecURL   string  `json:"api_spec_url"`
	DashboardURL string  `json:"dashboard_url"`
	RunbookURL   string  `json:"runbook_url"`
	SLOTarget    float64 `json:"slo_target"`
}

// AddEnvironmentRequest is the payload for registering a deployment environment on a service.
type AddEnvironmentRequest struct {
	Environment     string `json:"environment" binding:"required,max=50"`
	ClusterName     string `json:"cluster_name"`
	Namespace       string `json:"namespace"`
	ReplicasMin     int32  `json:"replicas_min"`
	ReplicasMax     int32  `json:"replicas_max"`
	CPURequest      string `json:"cpu_request"`
	MemoryRequest   string `json:"memory_request"`
	CurrentImageTag string `json:"current_image_tag"`
	URL             string `json:"url"`
}

// AddDependencyRequest is the payload for declaring that a service depends on another service.
type AddDependencyRequest struct {
	DependsOnServiceID string `json:"depends_on_service_id" binding:"required,uuid"`
	DependencyType     string `json:"dependency_type"`
	IsCritical         bool   `json:"is_critical"`
}

// AddTagRequest is the payload for tagging a service.
type AddTagRequest struct {
	Tag string `json:"tag" binding:"required,max=100"`
}

// Response is the public representation of a service.
type Response struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Description     string  `json:"description,omitempty"`
	ServiceTypeID   int16   `json:"service_type_id"`
	LifecycleID     int16   `json:"lifecycle_id,omitempty"`
	TierID          int16   `json:"tier_id,omitempty"`
	ProjectID       string  `json:"project_id,omitempty"`
	OwnerTeamID     string  `json:"owner_team_id"`
	TechLeadID      string  `json:"tech_lead_id,omitempty"`
	RepoURL         string  `json:"repo_url"`
	RepoProviderID  int16   `json:"repo_provider_id,omitempty"`
	DefaultBranch   string  `json:"default_branch"`
	LanguageID      int16   `json:"language_id,omitempty"`
	Framework       string  `json:"framework,omitempty"`
	DockerfilePath  string  `json:"dockerfile_path"`
	RegistryImage   string  `json:"registry_image,omitempty"`
	CIPipelineURL   string  `json:"ci_pipeline_url,omitempty"`
	GitopsRepoPath  string  `json:"gitops_repo_path,omitempty"`
	K8sResourceKind string  `json:"k8s_resource_kind"`
	Port            int32   `json:"port,omitempty"`
	HealthCheckPath string  `json:"health_check_path"`
	InternalURL     string  `json:"internal_url,omitempty"`
	ExternalURL     string  `json:"external_url,omitempty"`
	APISpecURL      string  `json:"api_spec_url,omitempty"`
	DashboardURL    string  `json:"dashboard_url,omitempty"`
	RunbookURL      string  `json:"runbook_url,omitempty"`
	SLOTarget       float64 `json:"slo_target,omitempty"`
	IsActive        bool    `json:"is_active"`
}

// EnvironmentResponse is the public representation of a service environment.
type EnvironmentResponse struct {
	ID              string `json:"id"`
	ServiceID       string `json:"service_id"`
	Environment     string `json:"environment"`
	ClusterName     string `json:"cluster_name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ReplicasMin     int32  `json:"replicas_min"`
	ReplicasMax     int32  `json:"replicas_max"`
	CPURequest      string `json:"cpu_request,omitempty"`
	MemoryRequest   string `json:"memory_request,omitempty"`
	CurrentImageTag string `json:"current_image_tag,omitempty"`
	URL             string `json:"url,omitempty"`
}

// DependencyResponse is the public representation of a service dependency edge.
type DependencyResponse struct {
	ServiceID          string `json:"service_id"`
	DependsOnServiceID string `json:"depends_on_service_id"`
	DependencyType     string `json:"dependency_type,omitempty"`
	IsCritical         bool   `json:"is_critical"`
}
