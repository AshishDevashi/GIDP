package lookup

import (
	"context"

	"github.com/AshishDevashi/GIDP/internal/platform/pgtext"
	"github.com/AshishDevashi/GIDP/internal/store"
)

// Service exposes read-only access to the static catalog lookup tables.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Lifecycles(ctx context.Context) ([]Item, error) {
	rows, err := s.repo.ListLifecycles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(rows))
	for i, r := range rows {
		items[i] = Item{ID: r.ID, Code: r.Code, Label: pgtext.To(r.Label)}
	}
	return items, nil
}

func (s *Service) Tiers(ctx context.Context) ([]TierItem, error) {
	rows, err := s.repo.ListTiers(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]TierItem, len(rows))
	for i, r := range rows {
		items[i] = TierItem{
			ID:           r.ID,
			Code:         r.Code,
			Description:  pgtext.To(r.Description),
			PagingPolicy: pgtext.To(r.PagingPolicy),
		}
	}
	return items, nil
}

func (s *Service) ServiceTypes(ctx context.Context) ([]Item, error) {
	rows, err := s.repo.ListServiceTypes(ctx)
	if err != nil {
		return nil, err
	}
	return toItems(rows, func(r store.ServiceType) Item { return Item{ID: r.ID, Code: r.Code} }), nil
}

func (s *Service) RepoProviders(ctx context.Context) ([]Item, error) {
	rows, err := s.repo.ListRepoProviders(ctx)
	if err != nil {
		return nil, err
	}
	return toItems(rows, func(r store.RepoProvider) Item { return Item{ID: r.ID, Code: r.Code} }), nil
}

func (s *Service) Languages(ctx context.Context) ([]Item, error) {
	rows, err := s.repo.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(rows))
	for i, r := range rows {
		items[i] = Item{ID: r.ID, Code: r.Code, Label: pgtext.To(r.Label)}
	}
	return items, nil
}

func (s *Service) RepoTemplates(ctx context.Context) ([]RepoTemplateItem, error) {
	rows, err := s.repo.ListRepoTemplates(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]RepoTemplateItem, len(rows))
	for i, r := range rows {
		items[i] = RepoTemplateItem{
			ID:            r.ID,
			Name:          r.Name,
			Slug:          r.Slug,
			TemplateOwner: r.TemplateOwner,
			TemplateRepo:  r.TemplateRepo,
		}
	}
	return items, nil
}

func toItems[T any](rows []T, convert func(T) Item) []Item {
	items := make([]Item, len(rows))
	for i, r := range rows {
		items[i] = convert(r)
	}
	return items
}
