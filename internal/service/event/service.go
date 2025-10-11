package event

import (
	"context"
	"fmt"
	"time"

	domain "go-api-kbt/internal/domain/event"
	repo "go-api-kbt/internal/repository/event"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	repo repo.Repository
	v    *validator.Validate
}

func NewService(r repo.Repository) *Service { return &Service{repo: r, v: validator.New()} }

type CreateInput struct {
	Name        string `json:"name" validate:"required,min=1,max=128"`
	Description string `json:"description" validate:"omitempty,max=1024"`
}

type UpdateInput struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=128"`
	Description *string `json:"description" validate:"omitempty,max=1024"`
}

type DTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (DTO, error) {
	if s.v == nil {
		s.v = validator.New()
	}
	if err := s.v.Struct(in); err != nil {
		return DTO{}, fmt.Errorf("validate input: %w", err)
	}

	e := &domain.Entity{Name: in.Name, Description: in.Description}
	if err := s.repo.Create(ctx, e); err != nil {
		return DTO{}, fmt.Errorf("create event: %w", err)
	}
	return toDTO(e), nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]DTO, error) {
	es, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]DTO, 0, len(es))
	for _, e := range es {
		copy := e
		out = append(out, toDTO(&copy))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id uint) (DTO, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return DTO{}, err
	}
	return toDTO(e), nil
}

func (s *Service) Update(ctx context.Context, id uint, in UpdateInput) (DTO, error) {
	if s.v == nil {
		s.v = validator.New()
	}
	if err := s.v.Struct(in); err != nil {
		return DTO{}, fmt.Errorf("validate input: %w", err)
	}

	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return DTO{}, err
	}
	if in.Name != nil {
		e.Name = *in.Name
	}
	if in.Description != nil {
		e.Description = *in.Description
	}
	if err := s.repo.Update(ctx, e); err != nil {
		return DTO{}, err
	}
	return toDTO(e), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }

func toDTO(e *domain.Entity) DTO {
	return DTO{ID: e.ID, Name: e.Name, Description: e.Description, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}
}
