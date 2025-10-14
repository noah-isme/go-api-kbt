package activity

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "go-api-kbt/internal/domain/activity"

	"gorm.io/gorm"
)

// GormRepository is a GORM implementation of the activity repository.
type GormRepository struct {
	db               *gorm.DB
	statementTimeout time.Duration
}

// NewGormRepository constructs a new repository using the supplied database
// handle.
func NewGormRepository(db *gorm.DB, statementTimeout time.Duration) *GormRepository {
	return &GormRepository{db: db, statementTimeout: statementTimeout}
}

func (r *GormRepository) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.statementTimeout <= 0 {
		r.statementTimeout = 5 * time.Second
	}
	return context.WithTimeout(ctx, r.statementTimeout)
}

// FindByID fetches an activity by identifier.
func (r *GormRepository) FindByID(ctx context.Context, id uint) (*domain.Activity, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var entity domain.Activity
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("activity not found") // Using a generic error for now
		}
		return nil, fmt.Errorf("get activity by id: %w", err)
	}
	return &entity, nil
}

// FindAll returns all activities.
func (r *GormRepository) FindAll(ctx context.Context, limit, offset int, query, sort string) ([]domain.Activity, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var activities []domain.Activity
	var total int64

	db := r.db.WithContext(ctx)

	if query != "" {
		db = db.Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if sort != "" {
		db = db.Order(sort)
	} else {
		db = db.Order("id asc")
	}

	db.Model(&domain.Activity{}).Count(&total)

	if limit > 0 {
		db = db.Limit(limit).Offset(offset)
	}

	if err := db.Find(&activities).Error; err != nil {
		return nil, 0, fmt.Errorf("find all activities: %w", err)
	}
	return activities, int(total), nil
}
