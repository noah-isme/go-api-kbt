package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "go-api-kbt/internal/domain/user"

	"gorm.io/gorm"
)

// GormRepository is a GORM implementation of the user repository.
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

// Create persists a new user entity.
func (r *GormRepository) Create(ctx context.Context, user *domain.Entity) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetByID fetches a user by identifier.
func (r *GormRepository) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var entity domain.Entity
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &entity, nil
}

// GetByEmail fetches a user using their email address.
func (r *GormRepository) GetByEmail(ctx context.Context, email string) (*domain.Entity, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	var entity domain.Entity
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &entity, nil
}

// List returns a paginated list of users.
func (r *GormRepository) List(ctx context.Context, limit, offset int) ([]domain.Entity, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if limit <= 0 {
		limit = 20
	}

	var users []domain.Entity
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("id asc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// Update modifies a user record.
func (r *GormRepository) Update(ctx context.Context, user *domain.Entity) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// Delete removes a user by identifier.
func (r *GormRepository) Delete(ctx context.Context, id uint) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	if err := r.db.WithContext(ctx).Delete(&domain.Entity{}, id).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
