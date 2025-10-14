package medaler

import (
	"context"
	"fmt"

	domain "go-api-kbt/internal/domain/medaler"
	repo "go-api-kbt/internal/repository/medaler"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-playground/validator/v10"
)

// Service orchestrates medaler domain operations.
type Service struct {
	repository repo.MedalerRepository
	validator  *validator.Validate
}

// NewService constructs a new service.
func NewService(repository repo.MedalerRepository) *Service {
	return &Service{
		repository: repository,
		validator:  validator.New(validator.WithRequiredStructEnabled()),
	}
}

// ListInput defines input for listing medalers with pagination and filtering.
type ListInput struct {
	Page  int
	Limit int
	Query string
	Sort  string
}

// MedalerListResponse represents the response for listing medalers.
type MedalerListResponse struct {
	Data []dto.MedalerDTO `json:"data"`
	Meta dto.PaginationMeta   `json:"meta"`
}

// Create registers a new medaler.
func (s *Service) Create(ctx context.Context, input dto.CreateMedalerInput) (dto.MedalerDTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.MedalerDTO{}, fmt.Errorf("validate create input: %w", err)
	}

	entity := &domain.Medaler{
		Name:     input.Name,
		Email:    input.Email,
		IsActive: true, // Default value
	}

	if err := s.repository.Create(ctx, entity); err != nil {
		return dto.MedalerDTO{}, fmt.Errorf("create medaler: %w", err)
	}

	return toDTO(entity), nil
}

// List returns all medalers.
func (s *Service) List(ctx context.Context, input ListInput) ([]dto.MedalerDTO, dto.PaginationMeta, error) {
	medalers, total, err := s.repository.FindAll(ctx, input.Limit, input.Page*input.Limit, input.Query, input.Sort)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("list medalers: %w", err)
	}

	dtos := make([]dto.MedalerDTO, 0, len(medalers))
	for _, m := range medalers {
		medaler := m
		dtos = append(dtos, toDTO(&medaler))
	}

	meta := dto.PaginationMeta{
		Page:        input.Page,
		Limit:       input.Limit,
		TotalRecords: int(total),
	}

	return dtos, meta, nil
}

// GetByID fetches a medaler by ID.
func (s *Service) GetByID(ctx context.Context, id uint) (dto.MedalerDTO, error) {
	entity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.MedalerDTO{}, fmt.Errorf("get medaler by ID: %w", err)
	}
	return toDTO(entity), nil
}

// Update modifies an existing medaler.
func (s *Service) Update(ctx context.Context, id uint, input dto.UpdateMedalerInput) (dto.MedalerDTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.MedalerDTO{}, fmt.Errorf("validate update input: %w", err)
	}

	entity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.MedalerDTO{}, err
	}

	if input.Name != nil {
		entity.Name = *input.Name
	}
	if input.Email != nil {
		entity.Email = *input.Email
	}
	if input.IsActive != nil {
		entity.IsActive = *input.IsActive
	}

	if err := s.repository.Update(ctx, entity); err != nil {
		return dto.MedalerDTO{}, err
	}
	return toDTO(entity), nil
}

// Delete removes a medaler.
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func toDTO(entity *domain.Medaler) dto.MedalerDTO {
	return dto.MedalerDTO{
		ID:       entity.ID,
		Name:     entity.Name,
		Email:    entity.Email,
		IsActive: entity.IsActive,
	}
}
