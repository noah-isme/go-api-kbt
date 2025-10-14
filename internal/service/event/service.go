package event

import (
	"context"
	"fmt"
	"time"

	domain "go-api-kbt/internal/domain/event"
	repo "go-api-kbt/internal/repository/event"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	repo repo.Repository
	v    *validator.Validate
}

func NewService(r repo.Repository) *Service { return &Service{repo: r, v: validator.New()} }

// EventListResponse represents the response for listing events.
type EventListResponse struct {
	Data []dto.EventDTO `json:"data"`
	Meta dto.PaginationMeta `json:"meta"`
}

func (s *Service) Create(ctx context.Context, in dto.CreateEventInput) (dto.EventDTO, error) {
	if s.v == nil {
		s.v = validator.New()
	}
	if err := s.v.Struct(in); err != nil {
		return dto.EventDTO{}, fmt.Errorf("validate input: %w", err)
	}

	e := &domain.Entity{Name: in.Name, Description: in.Description}
	if err := s.repo.Create(ctx, e); err != nil {
		return dto.EventDTO{}, fmt.Errorf("create event: %w", err)
	}
	return toDTO(e), nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]dto.EventDTO, error) {
	es, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]dto.EventDTO, 0, len(es))
	for _, e := range es {
		copy := e
		out = append(out, toDTO(&copy))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id uint) (dto.EventDTO, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.EventDTO{}, err
	}
	return toDTO(e), nil
}

func (s *Service) Update(ctx context.Context, id uint, in dto.UpdateEventInput) (dto.EventDTO, error) {
	if s.v == nil {
		s.v = validator.New()
	}
	if err := s.v.Struct(in); err != nil {
		return dto.EventDTO{}, fmt.Errorf("validate input: %w", err)
	}

	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return dto.EventDTO{}, err
	}
	if in.Name != nil {
		e.Name = *in.Name
	}
	if in.Description != nil {
		e.Description = *in.Description
	}
	if err := s.repo.Update(ctx, e); err != nil {
		return dto.EventDTO{}, err
	}
	return toDTO(e), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }

func toDTO(e *domain.Entity) dto.EventDTO {
	return dto.EventDTO{ID: e.ID, Name: e.Name, Description: e.Description, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}
}
