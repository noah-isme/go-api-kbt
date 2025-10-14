package dto

import (
	"time"

	"go-api-kbt/internal/domain/user"
)

// UserDTO represents the user as exposed to the transport layer.
type UserDTO struct {
	ID        uint        `json:"id"`
	Username  string      `json:"username"`
	Email     string      `json:"email"`
	Role      user.Role `json:"role"`
	EventID   *uint       `json:"event_id"`
	LastLogin *time.Time  `json:"last_login_at"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CreateUserInput contains the required fields for creating a user.
type CreateUserInput struct {
	Username string      `json:"username" validate:"required,min=3,max=64"`
	Email    string      `json:"email" validate:"required,email"`
	Password string      `json:"password" validate:"required,min=8"`
	Role     user.Role `json:"role" validate:"required,role"`
	EventID  *uint       `json:"event_id"`
}

// UpdateUserInput defines fields that can be updated.
type UpdateUserInput struct {
	Username *string      `json:"username" validate:"omitempty,min=3,max=64"`
	Email    *string      `json:"email" validate:"omitempty,email"`
	Password *string      `json:"password" validate:"omitempty,min=8"`
	Role     *user.Role `json:"role" validate:"omitempty,role"`
	EventID  *uint        `json:"event_id"`
}
