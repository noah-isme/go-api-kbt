package activity

import (
	"context"
	"fmt"

	domain "go-api-kbt/internal/domain/activity"
	repo "go-api-kbt/internal/repository/activity"
	"go-api-kbt/internal/transport/http/dto"
)

// Service orchestrates activity domain operations.
type Service struct {
	repository repo.ActivityRepository
}

// NewService constructs a new service.
func NewService(repository repo.ActivityRepository) *Service {
	return &Service{
		repository: repository,
	}
}

// ListInput defines input for listing activities with pagination and filtering.
type ListInput struct {
	Page  int
	Limit int
	Query string
	Sort  string
}

// ActivityListResponse represents the response for listing activities.
type ActivityListResponse struct {
	Data []dto.ActivityDTO  `json:"data"`
	Meta dto.PaginationMeta `json:"meta"`
}

// List returns all activities.
func (s *Service) List(ctx context.Context, input ListInput) ([]dto.ActivityDTO, dto.PaginationMeta, error) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.Limit <= 0 {
		input.Limit = 20
	}

	offset := (input.Page - 1) * input.Limit

	activities, total, err := s.repository.FindAll(ctx, input.Limit, offset, input.Query, input.Sort)
	if err != nil {
		return nil, dto.PaginationMeta{}, fmt.Errorf("list activities: %w", err)
	}

	dtos := make([]dto.ActivityDTO, 0, len(activities))
	for _, a := range activities {
		activity := a
		dtos = append(dtos, toDTO(&activity))
	}

	meta := dto.PaginationMeta{
		Page:         input.Page,
		Limit:        input.Limit,
		TotalRecords: int(total),
	}

	return dtos, meta, nil
}

// GetByID fetches an activity by ID.
func (s *Service) GetByID(ctx context.Context, id uint) (dto.ActivityDTO, error) {
	entity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return dto.ActivityDTO{}, fmt.Errorf("get activity by ID: %w", err)
	}
	return toDTO(entity), nil
}

func toDTO(entity *domain.Activity) dto.ActivityDTO {
	return dto.ActivityDTO{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
	}
}
