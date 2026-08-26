package team

// CreateTeamRequest is the payload for creating a new team.
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,max=150"`
	Slug        string `json:"slug" binding:"required,max=150"`
	Description string `json:"description"`
}

// AddMemberRequest is the payload for adding a user to a team.
type AddMemberRequest struct {
	UserID     string `json:"user_id" binding:"required,uuid"`
	RoleInTeam string `json:"role_in_team"`
	IsPrimary  bool   `json:"is_primary"`
}

// Response is the public representation of a team.
type Response struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// MemberResponse is the public representation of a team membership.
type MemberResponse struct {
	ID         string  `json:"id"`
	TeamID     string  `json:"team_id"`
	UserID     string  `json:"user_id"`
	RoleInTeam string  `json:"role_in_team"`
	IsPrimary  bool    `json:"is_primary"`
	LeftAt     *string `json:"left_at,omitempty"`
}
