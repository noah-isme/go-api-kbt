package activity

import (
	"context"

	domain "go-api-kbt/internal/domain/activity"
)

type ActivityRepository interface {
	FindAll(ctx context.Context, limit, offset int, query, sort string) ([]domain.Activity, int, error)
	FindByID(ctx context.Context, id uint) (*domain.Activity, error)
}