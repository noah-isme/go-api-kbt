package location

import (
	"context"
	"fmt"

	domain "go-api-kbt/internal/domain/location"
	repo "go-api-kbt/internal/repository/location"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-playground/validator/v10"
)

// Service orchestrates location domain operations.
type Service struct {
	repository repo.Repository
	validator  *validator.Validate
}

// NewService constructs a new service.
func NewService(repository repo.Repository) *Service {
	return &Service{repository: repository, validator: validator.New(validator.WithRequiredStructEnabled())}
}

// LocationListResponse represents the response for listing locations.
type LocationListResponse struct {
	Data []dto.LocationDTO  `json:"data"`
	Meta dto.PaginationMeta `json:"meta"`
}

// Create registers a new location.
func (s *Service) Create(ctx context.Context, input dto.CreateLocationInput) (dto.LocationDTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.LocationDTO{}, fmt.Errorf("validate location input: %w", err)
	}

	entity := &domain.Entity{
		Name:    input.Name,
		Address: input.Address,
	}

	if err := s.repository.Create(ctx, entity); err != nil {
		return dto.LocationDTO{}, fmt.Errorf("create location: %w", err)
	}

	return toDTO(entity), nil
}

// List returns paginated locations with metadata.
func (s *Service) List(ctx context.Context, params dto.PaginationParams) (dto.LocationListResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit
	locations, total, err := s.repository.List(ctx, params.Limit, offset)
	if err != nil {
		return dto.LocationListResponse{}, err
	}

	dtos := make([]dto.LocationDTO, 0, len(locations))
	for _, l := range locations {
		location := l
		dtos = append(dtos, toDTO(&location))
	}

	meta := dto.PaginationMeta{
		Page:         params.Page,
		Limit:        params.Limit,
		TotalRecords: total,
	}

	return dto.LocationListResponse{Data: dtos, Meta: meta}, nil
}

// Get fetches a location by ID.
func (s *Service) Get(ctx context.Context, id uint) (dto.LocationDTO, error) {
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return dto.LocationDTO{}, err
	}
	return toDTO(entity), nil
}

// Update modifies an existing location.
func (s *Service) Update(ctx context.Context, id uint, input dto.UpdateLocationInput) (dto.LocationDTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.LocationDTO{}, fmt.Errorf("validate location input: %w", err)
	}

	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return dto.LocationDTO{}, err
	}

	if input.Name != nil {
		entity.Name = *input.Name
	}
	if input.Address != nil {
		entity.Address = *input.Address
	}

	if err := s.repository.Update(ctx, entity); err != nil {
		return dto.LocationDTO{}, err
	}
	return toDTO(entity), nil
}

// Delete removes a location.
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func toDTO(entity *domain.Entity) dto.LocationDTO {
	return dto.LocationDTO{
		ID:        entity.ID,
		Name:      entity.Name,
		Address:   entity.Address,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
