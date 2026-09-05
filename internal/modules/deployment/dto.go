package deployment

import "encoding/json"

type CreateRequest struct {
	RepoID     string               `json:"repo_id" binding:"omitempty,uuid"`
	RegistryID string               `json:"registry_id" binding:"required,uuid"`
	ImageName  string               `json:"image_name" binding:"required,max=255"`
	ImageTag   string               `json:"image_tag" binding:"required,max=255"`
	Name       string               `json:"name" binding:"required,max=255"`
	Namespace  string               `json:"namespace" binding:"omitempty,max=255"`
	Replicas   int32                `json:"replicas" binding:"omitempty,min=1"`
	Resources  ResourceRequest      `json:"resources"`
	EnvVars    map[string]string    `json:"env_vars"`
	SecretRefs map[string]SecretRef `json:"secret_refs"`
	Expose     ExposeRequest        `json:"expose"`
}

type ResourceRequest struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type SecretRef struct {
	SecretName string `json:"secret_name"`
	Key        string `json:"key"`
}

type ExposeRequest struct {
	Type       string `json:"type"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"target_port"`
	Host       string `json:"host"`
	Path       string `json:"path"`
}

type Response struct {
	ID                   string          `json:"id"`
	DeploymentInstanceID string          `json:"deployment_instance_id"`
	RepoID               string          `json:"repo_id,omitempty"`
	RegistryID           string          `json:"registry_id"`
	ImageName            string          `json:"image_name"`
	ImageTag             string          `json:"image_tag"`
	ImageRef             string          `json:"image_ref,omitempty"`
	Name                 string          `json:"name"`
	Namespace            string          `json:"namespace"`
	Replicas             int32           `json:"replicas"`
	Resources            json.RawMessage `json:"resources"`
	EnvVars              json.RawMessage `json:"env_vars"`
	SecretRefs           json.RawMessage `json:"secret_refs"`
	Expose               json.RawMessage `json:"expose"`
	Status               string          `json:"status"`
	CurrentRevision      int32           `json:"current_revision"`
	K8sDeploymentName    string          `json:"k8s_deployment_name,omitempty"`
	LastError            string          `json:"last_error,omitempty"`
	LatestRevision       *Revision       `json:"latest_revision,omitempty"`
	K8sStatus            *K8sStatus      `json:"k8s_status,omitempty"`
	CreatedBy            string          `json:"created_by"`
	CreatedAt            string          `json:"created_at"`
	UpdatedAt            string          `json:"updated_at"`
}

type Revision struct {
	ID        string          `json:"id"`
	Number    int32           `json:"number"`
	ImageTag  string          `json:"image_tag"`
	Snapshot  json.RawMessage `json:"config_snapshot,omitempty"`
	Status    string          `json:"status"`
	CreatedAt string          `json:"created_at"`
}

type K8sStatus struct {
	State               string         `json:"state"`
	Replicas            int32          `json:"replicas"`
	ReadyReplicas       int32          `json:"ready_replicas"`
	AvailableReplicas   int32          `json:"available_replicas"`
	UnavailableReplicas int32          `json:"unavailable_replicas"`
	Pods                int            `json:"pods"`
	PodPhases           map[string]int `json:"pod_phases,omitempty"`
	Error               string         `json:"error,omitempty"`
}
