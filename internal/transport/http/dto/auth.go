package dto

// LoginInput defines the input for user login.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenInput defines the input for refreshing tokens.
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthOutput defines the output for authentication.
type AuthOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
