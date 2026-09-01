package repo

// CreateRequest is the payload for creating a GitHub repository.
type CreateRequest struct {
	Name         string `json:"name" binding:"required,max=100"`
	Description  string `json:"description" binding:"max=350"`
	Organization string `json:"organization" binding:"omitempty,max=39"`
	Private      bool   `json:"private"`
	AutoInit     bool   `json:"auto_init"`
}

// UpdateRequest is the payload for updating repository metadata in the portal.
type UpdateRequest struct {
	Name          string `json:"name" binding:"required,max=150"`
	DefaultBranch string `json:"default_branch" binding:"required,max=100"`
	Visibility    string `json:"visibility" binding:"required,oneof=public private internal"`
	TemplateUsed  string `json:"template_used" binding:"max=150"`
	Status        string `json:"status" binding:"required,oneof=pending creating active failed archived"`
}

// Response is the public representation of a persisted repository.
type Response struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name,omitempty"`
	Owner         string `json:"owner"`
	ProviderID    int16  `json:"provider_id"`
	ExternalID    string `json:"external_id,omitempty"`
	URL           string `json:"url,omitempty"`
	CloneURLSSH   string `json:"clone_url_ssh,omitempty"`
	CloneURLHTTPS string `json:"clone_url_https,omitempty"`
	DefaultBranch string `json:"default_branch"`
	Visibility    string `json:"visibility"`
	TemplateUsed  string `json:"template_used,omitempty"`
	Status        string `json:"status"`
	ErrorMessage  string `json:"error_message,omitempty"`
	CreatedBy     string `json:"created_by,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
