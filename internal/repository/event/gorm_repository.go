package event

import (
	"context"
	"errors"
	"fmt"

	domain "go-api-kbt/internal/domain/event"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, e *domain.Entity) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (r *GormRepository) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
	var e domain.Entity
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get event: %w", err)
	}
	return &e, nil
}

func (r *GormRepository) List(ctx context.Context, limit, offset int) ([]domain.Entity, error) {
	if limit <= 0 {
		limit = 20
	}
	var out []domain.Entity
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("id asc").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return out, nil
}

func (r *GormRepository) Update(ctx context.Context, e *domain.Entity) error {
	if err := r.db.WithContext(ctx).Save(e).Error; err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (r *GormRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Entity{}, id).Error; err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}
