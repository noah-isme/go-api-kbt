package dto

import "time"

// EventDTO represents the event as exposed to the transport layer.
type EventDTO struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateEventInput contains the required fields for creating an event.
type CreateEventInput struct {
	Name        string `json:"name" validate:"required,min=1,max=128"`
	Description string `json:"description" validate:"omitempty,max=1024"`
}

// UpdateEventInput defines fields that can be updated.
type UpdateEventInput struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=128"`
	Description *string `json:"description" validate:"omitempty,max=1024"`
}
