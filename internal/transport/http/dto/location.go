package dto

import "time"

// LocationDTO represents the location as exposed to the transport layer.
type LocationDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateLocationInput contains the required fields for creating a location.
type CreateLocationInput struct {
	Name    string `json:"name" validate:"required,min=3,max=128"`
	Address string `json:"address" validate:"required,min=3,max=255"`
}

// UpdateLocationInput defines fields that can be updated.
type UpdateLocationInput struct {
	Name    *string `json:"name" validate:"omitempty,min=3,max=128"`
	Address *string `json:"address" validate:"omitempty,min=3,max=255"`
}
