package location

import (
	"context"
	"fmt"
	"time"

	domain "go-api-kbt/internal/domain/location"
	repo "go-api-kbt/internal/repository/location"
	"go-api-kbt/internal/transport/http/dto"
)

// Service orchestrates location domain operations.
type Service struct {
	repository repo.Repository
}

// NewService constructs a new service.
func NewService(repository repo.Repository) *Service {
	return &Service{repository: repository}
}

// LocationListResponse represents the response for listing locations.
type LocationListResponse struct {
	Data []dto.LocationDTO `json:"data"`
	Meta dto.PaginationMeta    `json:"meta"`
}

// Create registers a new location.
func (s *Service) Create(ctx context.Context, input dto.CreateLocationInput) (dto.LocationDTO, error) {
	entity := &domain.Entity{
		Name:    input.Name,
		Address: input.Address,
	}

	if err := s.repository.Create(ctx, entity); err != nil {
		return dto.LocationDTO{}, fmt.Errorf("create location: %w", err)
	}

	return toDTO(entity), nil
}

// List returns paginated locations.
func (s *Service) List(ctx context.Context, limit, offset int) ([]dto.LocationDTO, error) {
	locations, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.LocationDTO, 0, len(locations))
	for _, l := range locations {
		location := l
		dtos = append(dtos, toDTO(&location))
	}
	return dtos, nil
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
