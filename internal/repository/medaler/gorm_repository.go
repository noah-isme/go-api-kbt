package medaler

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "go-api-kbt/internal/domain/medaler"

	"gorm.io/gorm"
)

// GormRepository is a GORM implementation of the medaler repository.
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

// Create persists a new medaler entity.
func (r *GormRepository) Create(ctx context.Context, medaler *domain.Medaler) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if err := r.db.WithContext(ctx).Create(medaler).Error; err != nil {
		return fmt.Errorf("create medaler: %w", err)
	}
	return nil
}

// FindByID fetches a medaler by identifier.
func (r *GormRepository) FindByID(ctx context.Context, id uint) (*domain.Medaler, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var entity domain.Medaler
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("medaler not found") // Using a generic error for now
		}
		return nil, fmt.Errorf("get medaler by id: %w", err)
	}
	return &entity, nil
}

// FindAll returns all medalers.
func (r *GormRepository) FindAll(ctx context.Context, limit, offset int, query, sort string) ([]domain.Medaler, int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var medalers []domain.Medaler
	var total int64

	db := r.db.WithContext(ctx)

	if query != "" {
		db = db.Where("name ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if sort != "" {
		db = db.Order(sort)
	} else {
		db = db.Order("id asc")
	}

	db.Model(&domain.Medaler{}).Count(&total)

	if limit > 0 {
		db = db.Limit(limit).Offset(offset)
	}

	if err := db.Find(&medalers).Error; err != nil {
		return nil, 0, fmt.Errorf("find all medalers: %w", err)
	}
	return medalers, int(total), nil
}

// Update modifies a medaler record.
func (r *GormRepository) Update(ctx context.Context, medaler *domain.Medaler) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if err := r.db.WithContext(ctx).Save(medaler).Error; err != nil {
		return fmt.Errorf("update medaler: %w", err)
	}
	return nil
}

// Delete removes a medaler by identifier.
func (r *GormRepository) Delete(ctx context.Context, id uint) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if err := r.db.WithContext(ctx).Delete(&domain.Medaler{}, id).Error; err != nil {
		return fmt.Errorf("delete medaler: %w", err)
	}
	return nil
}
