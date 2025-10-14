package auth

import (
	"context"
	"fmt"
	"time"

	"go-api-kbt/internal/config"
	userDomain "go-api-kbt/internal/domain/user"
	userRepo "go-api-kbt/internal/repository/user"
	authRepo "go-api-kbt/internal/repository/auth"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUnauthorized       = fmt.Errorf("unauthorized")
)

// Service provides authentication related business logic.
type Service struct {
	userRepository userRepo.Repository
	authRepository authRepo.Repository
	validator      *validator.Validate
	cfg            *config.AuthConfig
}

// NewService creates a new authentication service.
func NewService(userRepository userRepo.Repository, authRepository authRepo.Repository, cfg *config.AuthConfig) *Service {
	return &Service{
		userRepository: userRepository,
		authRepository: authRepository,
		validator:      validator.New(validator.WithRequiredStructEnabled()),
		cfg:            cfg,
	}
}

// Claims defines the JWT claims.
type Claims struct {
	UserID uint        `json:"user_id"`
	Role   userDomain.Role `json:"role"`
	jwt.RegisteredClaims
}

// Login authenticates a user and returns JWT tokens.
func (s *Service) Login(ctx context.Context, input dto.LoginInput) (dto.AuthOutput, error) {
	if err := s.validator.Struct(input); err != nil {
		return dto.AuthOutput{}, fmt.Errorf("validate login input: %w", err)
	}

	user, err := s.userRepository.GetByEmail(ctx, input.Email)
	if err != nil {
		return dto.AuthOutput{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return dto.AuthOutput{}, ErrInvalidCredentials
	}

	return s.generateTokens(ctx, user)
}

// Refresh generates new access and refresh tokens from a valid refresh token.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (dto.AuthOutput, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return dto.AuthOutput{}, ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Subject == "access" {
		return dto.AuthOutput{}, ErrUnauthorized
	}

	// Check if refresh token is in Redis
	if _, err := s.authRepository.GetRefreshToken(ctx, claims.UserID, claims.ID); err != nil {
		return dto.AuthOutput{}, ErrUnauthorized
	}

	// Delete old refresh token
	if err := s.authRepository.DeleteRefreshToken(ctx, claims.UserID, claims.ID); err != nil {
		return dto.AuthOutput{}, fmt.Errorf("delete old refresh token: %w", err)
	}

	user, err := s.userRepository.GetByID(ctx, claims.UserID)
	if err != nil {
		return dto.AuthOutput{}, ErrUnauthorized
	}

	return s.generateTokens(ctx, user)
}

// Logout invalidates a refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil // Already logged out or invalid token, no error
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Subject == "access" {
		return nil // Not a refresh token, no error
	}

	return s.authRepository.DeleteRefreshToken(ctx, claims.UserID, claims.ID)
}

func (s *Service) generateTokens(ctx context.Context, user *userDomain.Entity) (dto.AuthOutput, error) {
	accessTokenClaims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "access",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return dto.AuthOutput{}, fmt.Errorf("sign access token: %w", err)
	}

	refreshTokenID := uuid.New().String()
	refreshTokenClaims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        refreshTokenID,
			Subject:   "refresh",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return dto.AuthOutput{}, fmt.Errorf("sign refresh token: %w", err)
	}

	if err := s.authRepository.StoreRefreshToken(ctx, user.ID, refreshTokenID, s.cfg.RefreshTokenTTL); err != nil {
		return dto.AuthOutput{}, fmt.Errorf("store refresh token: %w", err)
	}

	return dto.AuthOutput{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	},
	nil
}