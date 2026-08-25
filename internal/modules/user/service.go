package user

import "context"

// Service contains the user module's business logic.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) (any, error) {
	return s.repo.List(ctx)
}
