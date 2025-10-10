package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "go-api-kbt/internal/domain/user"
	repo "go-api-kbt/internal/repository/user"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrEmailAlreadyExists indicates a conflict when creating a user with an
	// existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials indicates authentication failure.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Service orchestrates user domain operations.
type Service struct {
	repository repo.Repository
	validator  *validator.Validate
	hashCost   int
}

// NewService constructs a new service.
func NewService(repository repo.Repository, hashCost int) *Service {
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.RegisterValidation("role", func(fl validator.FieldLevel) bool {
		value := domain.Role(strings.ToLower(fl.Field().String()))
		switch value {
		case domain.RoleOwner, domain.RoleAdmin, domain.RoleCashier:
			return true
		default:
			return false
		}
	}); err != nil {
		panic(fmt.Sprintf("failed to register role validation: %v", err))
	}

	return &Service{
		repository: repository,
		validator:  v,
		hashCost:   hashCost,
	}
}

// CreateInput contains the required fields for creating a user.
type CreateInput struct {
	Username string      `json:"username" validate:"required,min=3,max=64"`
	Email    string      `json:"email" validate:"required,email"`
	Password string      `json:"password" validate:"required,min=8"`
	Role     domain.Role `json:"role" validate:"required,role"`
	EventID  *uint       `json:"event_id"`
}

// UpdateInput defines fields that can be updated.
type UpdateInput struct {
	Username *string      `json:"username" validate:"omitempty,min=3,max=64"`
	Email    *string      `json:"email" validate:"omitempty,email"`
	Password *string      `json:"password" validate:"omitempty,min=8"`
	Role     *domain.Role `json:"role" validate:"omitempty,role"`
	EventID  *uint        `json:"event_id"`
}

// DTO represents the user as exposed to the transport layer.
type DTO struct {
	ID        uint        `json:"id"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	EventID   *uint       `json:"event_id"`
	LastLogin *time.Time  `json:"last_login_at"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// AuthenticateInput contains login credentials.
type AuthenticateInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthenticateOutput holds authentication result details.
type AuthenticateOutput struct {
	User  DTO    `json:"user"`
	Token string `json:"access_token"`
}

// Create registers a new user.
func (s *Service) Create(ctx context.Context, input CreateInput) (DTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return DTO{}, fmt.Errorf("validate create input: %w", err)
	}

	if existing, err := s.repository.GetByEmail(ctx, strings.ToLower(input.Email)); err == nil && existing != nil {
		return DTO{}, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return DTO{}, fmt.Errorf("check existing user: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.hashCost)
	if err != nil {
		return DTO{}, fmt.Errorf("hash password: %w", err)
	}

	entity := &domain.Entity{
		Username: input.Username,
		Email:    strings.ToLower(input.Email),
		Password: string(hashed),
		Role:     input.Role,
		EventID:  input.EventID,
	}

	if err := s.repository.Create(ctx, entity); err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return DTO{}, ErrEmailAlreadyExists
		}
		return DTO{}, err
	}

	return toDTO(entity), nil
}

// List returns paginated users.
func (s *Service) List(ctx context.Context, limit, offset int) ([]DTO, error) {
	users, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]DTO, 0, len(users))
	for _, u := range users {
		user := u
		dtos = append(dtos, toDTO(&user))
	}
	return dtos, nil
}

// Get fetches a user by ID.
func (s *Service) Get(ctx context.Context, id uint) (DTO, error) {
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return DTO{}, err
	}
	return toDTO(entity), nil
}

// Update modifies an existing user.
func (s *Service) Update(ctx context.Context, id uint, input UpdateInput) (DTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return DTO{}, fmt.Errorf("validate update input: %w", err)
	}

	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return DTO{}, err
	}

	if input.Username != nil {
		entity.Username = *input.Username
	}
	if input.Email != nil {
		entity.Email = strings.ToLower(*input.Email)
	}
	if input.Role != nil {
		entity.Role = *input.Role
	}
	if input.EventID != nil {
		entity.EventID = input.EventID
	}
	if input.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), s.hashCost)
		if err != nil {
			return DTO{}, fmt.Errorf("hash password: %w", err)
		}
		entity.Password = string(hashed)
	}

	if err := s.repository.Update(ctx, entity); err != nil {
		return DTO{}, err
	}
	return toDTO(entity), nil
}

// Delete removes a user.
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func toDTO(entity *domain.Entity) DTO {
	return DTO{
		ID:        entity.ID,
		Username:  entity.Username,
		Email:     entity.Email,
		Role:      entity.Role,
		EventID:   entity.EventID,
		LastLogin: entity.LastLoginAt,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
