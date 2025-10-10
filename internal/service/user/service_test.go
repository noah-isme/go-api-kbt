package user_test

import (
	"context"
	"errors"
	"testing"

	domain "go-api-kbt/internal/domain/user"
	repo "go-api-kbt/internal/repository/user"
	"go-api-kbt/internal/service/user"
)

type mockRepository struct {
	createFn     func(ctx context.Context, entity *domain.Entity) error
	getByEmailFn func(ctx context.Context, email string) (*domain.Entity, error)
}

func (m *mockRepository) Create(ctx context.Context, entity *domain.Entity) error {
	if m.createFn != nil {
		return m.createFn(ctx, entity)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uint) (*domain.Entity, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*domain.Entity, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, repo.ErrNotFound
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]domain.Entity, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepository) Update(ctx context.Context, entity *domain.Entity) error {
	return errors.New("not implemented")
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	return errors.New("not implemented")
}

var _ repo.Repository = (*mockRepository)(nil)

func TestServiceCreateValidation(t *testing.T) {
	repository := &mockRepository{}
	service := user.NewService(repository, 4)

	_, err := service.Create(context.Background(), user.CreateInput{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestServiceCreateSuccess(t *testing.T) {
	repository := &mockRepository{
		getByEmailFn: func(ctx context.Context, email string) (*domain.Entity, error) {
			return nil, repo.ErrNotFound
		},
		createFn: func(ctx context.Context, entity *domain.Entity) error {
			entity.ID = 1
			return nil
		},
	}

	service := user.NewService(repository, 4)

	dto, err := service.Create(context.Background(), user.CreateInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     domain.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dto.ID != 1 {
		t.Fatalf("expected ID 1, got %d", dto.ID)
	}
	if dto.Email != "test@example.com" {
		t.Fatalf("expected email to be preserved")
	}
}
