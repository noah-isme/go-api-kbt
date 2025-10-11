package location

import (
    "context"
    "fmt"
    "time"

    domain "go-api-kbt/internal/domain/location"
    repo "go-api-kbt/internal/repository/location"

    "github.com/go-playground/validator/v10"
)

type Service struct{ repo repo.Repository; v *validator.Validate }

func NewService(r repo.Repository) *Service { return &Service{repo: r, v: validator.New()} }

type CreateInput struct {
    UserID    uint    `json:"user_id" validate:"required"`
    EventID   uint    `json:"event_id" validate:"required"`
    Latitude  float64 `json:"latitude" validate:"required"`
    Longitude float64 `json:"longitude" validate:"required"`
    Timestamp int64   `json:"timestamp" validate:"required"`
}

type UpdateInput struct {
    Latitude  *float64 `json:"latitude" validate:"omitempty"`
    Longitude *float64 `json:"longitude" validate:"omitempty"`
    Timestamp *int64   `json:"timestamp" validate:"omitempty"`
}

type DTO struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    EventID   uint      `json:"event_id"`
    Latitude  float64   `json:"latitude"`
    Longitude float64   `json:"longitude"`
    Timestamp int64     `json:"timestamp"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (DTO, error) {
    if s.v == nil {
        s.v = validator.New()
    }
    if err := s.v.Struct(in); err != nil {
        return DTO{}, fmt.Errorf("validate input: %w", err)
    }

    e := &domain.Entity{UserID: in.UserID, EventID: in.EventID, Latitude: in.Latitude, Longitude: in.Longitude, Timestamp: in.Timestamp}
    if err := s.repo.Create(ctx, e); err != nil {
        return DTO{}, fmt.Errorf("create location: %w", err)
    }
    return toDTO(e), nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]DTO, error) {
    ls, err := s.repo.List(ctx, limit, offset)
    if err != nil {
        return nil, err
    }
    out := make([]DTO, 0, len(ls))
    for _, l := range ls {
        copy := l
        out = append(out, toDTO(&copy))
    }
    return out, nil
}

func (s *Service) Get(ctx context.Context, id uint) (DTO, error) {
    l, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return DTO{}, err
    }
    return toDTO(l), nil
}

func (s *Service) Update(ctx context.Context, id uint, in UpdateInput) (DTO, error) {
    if s.v == nil {
        s.v = validator.New()
    }
    if err := s.v.Struct(in); err != nil {
        return DTO{}, fmt.Errorf("validate input: %w", err)
    }

    l, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return DTO{}, err
    }
    if in.Latitude != nil { l.Latitude = *in.Latitude }
    if in.Longitude != nil { l.Longitude = *in.Longitude }
    if in.Timestamp != nil { l.Timestamp = *in.Timestamp }
    if err := s.repo.Update(ctx, l); err != nil { return DTO{}, err }
    return toDTO(l), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }

func toDTO(l *domain.Entity) DTO {
    return DTO{ID: l.ID, UserID: l.UserID, EventID: l.EventID, Latitude: l.Latitude, Longitude: l.Longitude, Timestamp: l.Timestamp, CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt}
}
