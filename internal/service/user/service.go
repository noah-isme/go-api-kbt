package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain "go-api-kbt/internal/domain/user"
	repo "go-api-kbt/internal/repository/user"
	"go-api-kbt/internal/transport/http/dto"

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

// UserListResponse represents the response for listing users.
type UserListResponse struct {
	Data []dto.UserDTO      `json:"data"`
	Meta dto.PaginationMeta `json:"meta"`
}

// AuthenticateInput contains login credentials.
type AuthenticateInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthenticateOutput holds authentication result details.
type AuthenticateOutput struct {
	User  dto.UserDTO `json:"user"`
	Token string      `json:"access_token"`
}

// Create registers a new user.
func (s *Service) Create(ctx context.Context, input dto.CreateUserInput) (dto.UserDTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.UserDTO{}, fmt.Errorf("validate create input: %w", err)
	}

	if existing, err := s.repository.GetByEmail(ctx, strings.ToLower(input.Email)); err == nil && existing != nil {
		return dto.UserDTO{}, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return dto.UserDTO{}, fmt.Errorf("check existing user: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), s.hashCost)
	if err != nil {
		return dto.UserDTO{}, fmt.Errorf("hash password: %w", err)
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
			return dto.UserDTO{}, ErrEmailAlreadyExists
		}
		return dto.UserDTO{}, err
	}

	return toDTO(entity), nil
}

// List returns paginated users along with pagination metadata.
func (s *Service) List(ctx context.Context, params dto.PaginationParams) (dto.UserListResponse, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit
	users, total, err := s.repository.List(ctx, params.Limit, offset)
	if err != nil {
		return dto.UserListResponse{}, err
	}

	dtos := make([]dto.UserDTO, 0, len(users))
	for _, u := range users {
		user := u
		dtos = append(dtos, toDTO(&user))
	}

	meta := dto.PaginationMeta{
		Page:         params.Page,
		Limit:        params.Limit,
		TotalRecords: total,
	}

	return dto.UserListResponse{Data: dtos, Meta: meta}, nil
}

// Get fetches a user by ID.
func (s *Service) Get(ctx context.Context, id uint) (dto.UserDTO, error) {
	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return dto.UserDTO{}, err
	}
	return toDTO(entity), nil
}

// Update modifies an existing user.
func (s *Service) Update(ctx context.Context, id uint, input dto.UpdateUserInput) (dto.UserDTO, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.UserDTO{}, fmt.Errorf("validate update input: %w", err)
	}

	entity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return dto.UserDTO{}, err
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
			return dto.UserDTO{}, fmt.Errorf("hash password: %w", err)
		}
		entity.Password = string(hashed)
	}

	if err := s.repository.Update(ctx, entity); err != nil {
		return dto.UserDTO{}, err
	}
	return toDTO(entity), nil
}

// Delete removes a user.
func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func toDTO(entity *domain.Entity) dto.UserDTO {
	return dto.UserDTO{
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
