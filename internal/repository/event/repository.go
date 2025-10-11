package event

import (
	"context"
	"errors"

	domain "go-api-kbt/internal/domain/event"
)

var ErrNotFound = errors.New("event not found")

type Repository interface {
	Create(ctx context.Context, e *domain.Entity) error
	GetByID(ctx context.Context, id uint) (*domain.Entity, error)
	List(ctx context.Context, limit, offset int) ([]domain.Entity, error)
	Update(ctx context.Context, e *domain.Entity) error
	Delete(ctx context.Context, id uint) error
}
