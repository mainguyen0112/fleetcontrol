package satellite

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, sat *Satellite) (*Satellite, error) {
	sat.Status = "Pending"
	sat.ManagedBy = "manual"
	return s.repo.Create(ctx, sat)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Satellite, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Satellite, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, region string) (*Satellite, error) {
	return s.repo.Update(ctx, id, region)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
