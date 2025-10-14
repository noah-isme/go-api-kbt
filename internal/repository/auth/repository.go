package auth

import (
	"context"
	"time"
)

// Repository defines the interface for authentication data operations.
type Repository interface {
	StoreRefreshToken(ctx context.Context, userID uint, tokenID string, expiresIn time.Duration) error
	GetRefreshToken(ctx context.Context, userID uint, tokenID string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error
	DeleteUserRefreshTokens(ctx context.Context, userID uint) error
}