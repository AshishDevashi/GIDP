package registry

// CreateRequest is the payload for creating a Docker Hub repository (registry).
type CreateRequest struct {
	Name            string `json:"name" binding:"required,max=255"`
	Namespace       string `json:"namespace" binding:"omitempty,max=255"`
	Description     string `json:"description" binding:"max=100"`
	FullDescription string `json:"full_description" binding:"max=25000"`
	Private         bool   `json:"private"`
}

// UpdateRequest is the payload for updating registry metadata in the portal.
type UpdateRequest struct {
	Description string `json:"description" binding:"max=255"`
	Visibility  string `json:"visibility" binding:"required,oneof=public private"`
	Status      string `json:"status" binding:"required,oneof=active failed archived"`
}

// Response is the public representation of a persisted registry.
type Response struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProviderID  int16  `json:"provider_id"`
	Namespace   string `json:"namespace"`
	FullName    string `json:"full_name"`
	RegistryURL string `json:"registry_url,omitempty"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
	URL         string `json:"url"`
	PullCommand string `json:"pull_command"`
	CreatedBy   string `json:"created_by,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
