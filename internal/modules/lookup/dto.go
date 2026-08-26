package lookup

// Item is the public representation of a simple lookup/reference value.
type Item struct {
	ID    int16  `json:"id"`
	Code  string `json:"code"`
	Label string `json:"label,omitempty"`
}

// TierItem is the public representation of a project tier, including its paging policy.
type TierItem struct {
	ID           int16  `json:"id"`
	Code         string `json:"code"`
	Description  string `json:"description,omitempty"`
	PagingPolicy string `json:"paging_policy,omitempty"`
}
