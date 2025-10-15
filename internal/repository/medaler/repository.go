package medaler

import (
	"context"

	"go-api-kbt/internal/domain/medaler"
)

type MedalerRepository interface {
	FindAll(ctx context.Context, limit, offset int, query, sort string) ([]medaler.Medaler, int, error)
	FindByID(ctx context.Context, id uint) (*medaler.Medaler, error)
	Create(ctx context.Context, m *medaler.Medaler) error
	Update(ctx context.Context, m *medaler.Medaler) error
	Delete(ctx context.Context, id uint) error
}
