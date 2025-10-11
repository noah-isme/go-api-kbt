package location

import (
	"context"
	"errors"

	domain "go-api-kbt/internal/domain/location"
)

var ErrNotFound = errors.New("location not found")

type Repository interface {
	Create(ctx context.Context, e *domain.Entity) error
	GetByID(ctx context.Context, id uint) (*domain.Entity, error)
	List(ctx context.Context, limit, offset int) ([]domain.Entity, error)
	Update(ctx context.Context, e *domain.Entity) error
	Delete(ctx context.Context, id uint) error
}
