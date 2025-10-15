package user

import (
	"context"
	"errors"

	domain "go-api-kbt/internal/domain/user"
)

// ErrNotFound is returned when a user cannot be located in the repository.
var ErrNotFound = errors.New("user not found")

// Repository defines the behaviours required to persist users.
type Repository interface {
	Create(ctx context.Context, user *domain.Entity) error
	GetByID(ctx context.Context, id uint) (*domain.Entity, error)
	GetByEmail(ctx context.Context, email string) (*domain.Entity, error)
	List(ctx context.Context, limit, offset int) ([]domain.Entity, int, error)
	Update(ctx context.Context, user *domain.Entity) error
	Delete(ctx context.Context, id uint) error
}
