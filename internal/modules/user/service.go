package user

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/platform/uuidutil"
	"github.com/AshishDevashi/GIDP/internal/store"
)

// Response is the public, password-free representation of a user.
type Response struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name,omitempty"`
	RoleID   string `json:"role_id"`
	IsActive bool   `json:"is_active"`
}

// Service contains the user module's business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]Response, len(users))
	for i, u := range users {
		resp[i] = toResponse(u)
	}
	return resp, nil
}

func toResponse(u store.User) Response {
	return Response{
		ID:       uuidutil.String(u.ID),
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName.String,
		RoleID:   uuidutil.String(u.RoleID),
		IsActive: u.IsActive,
	}
}
