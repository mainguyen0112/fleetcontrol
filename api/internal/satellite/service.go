package satellite

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrManagedByOperator = errors.New("satellite is managed by operator and cannot be edited manually")

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

func (s *Service) Update(ctx context.Context, id uuid.UUID, region string, fromOperator bool) (*Satellite, error) {
	sat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sat == nil {
		return nil, nil
	}
	if sat.ManagedBy == "operator" && !fromOperator {
		return nil, ErrManagedByOperator
	}
	return s.repo.Update(ctx, id, region)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, fromOperator bool) error {
	sat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if sat == nil {
		return nil
	}
	if sat.ManagedBy == "operator" && !fromOperator {
		return ErrManagedByOperator
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Heartbeat(ctx context.Context, id uuid.UUID) (*Satellite, error) {
	return s.repo.UpdateHeartbeat(ctx, id)
}
