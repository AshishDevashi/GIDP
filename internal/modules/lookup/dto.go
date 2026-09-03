package lookup

// Item is the public representation of a simple lookup/reference value.
type Item struct {
	ID    int16  `json:"id"`
	Code  string `json:"code"`
	Label string `json:"label,omitempty"`
}

// RepoTemplateItem is the public representation of a predefined repository template.
type RepoTemplateItem struct {
	ID            int16  `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	TemplateOwner string `json:"template_owner"`
	TemplateRepo  string `json:"template_repo"`
}

// AllLookupsResponse is the public payload for all lookup tables in a single API response.
type AllLookupsResponse struct {
	RepoProviders []Item             `json:"repo_providers"`
	Languages     []Item             `json:"languages"`
	RepoTemplates []RepoTemplateItem `json:"repo_templates"`
}
