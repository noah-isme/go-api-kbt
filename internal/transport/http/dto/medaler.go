package dto

// MedalerDTO represents the medaler as exposed to the transport layer.
type MedalerDTO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// CreateMedalerInput contains the required fields for creating a medaler.
type CreateMedalerInput struct {
	Name  string `json:"name" validate:"required,min=3,max=64"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateMedalerInput defines fields that can be updated.
type UpdateMedalerInput struct {
	Name     *string `json:"name" validate:"omitempty,min=3,max=64"`
	Email    *string `json:"email" validate:"omitempty,email"`
	IsActive *bool   `json:"is_active"`
}
