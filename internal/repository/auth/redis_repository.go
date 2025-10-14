package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository is a Redis implementation of the auth repository.
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository constructs a new Redis repository.
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

// StoreRefreshToken stores a refresh token in Redis.
func (r *RedisRepository) StoreRefreshToken(ctx context.Context, userID uint, tokenID string, expiresIn time.Duration) error {
	key := fmt.Sprintf("refresh:%d:%s", userID, tokenID)
	if err := r.client.Set(ctx, key, "1", expiresIn).Err(); err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken retrieves a refresh token from Redis.
func (r *RedisRepository) GetRefreshToken(ctx context.Context, userID uint, tokenID string) (string, error) {
	key := fmt.Sprintf("refresh:%d:%s", userID, tokenID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("refresh token not found")
		}
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return val, nil
}

// DeleteRefreshToken deletes a refresh token from Redis.
func (r *RedisRepository) DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error {
	key := fmt.Sprintf("refresh:%d:%s", userID, tokenID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

// DeleteUserRefreshTokens deletes all refresh tokens for a given user.
func (r *RedisRepository) DeleteUserRefreshTokens(ctx context.Context, userID uint) error {
	pattern := fmt.Sprintf("refresh:%d:*", userID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("get user refresh token keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("delete user refresh tokens: %w", err)
		}
	}
	return nil
}
