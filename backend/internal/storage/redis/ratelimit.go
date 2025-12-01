package redis

import (
	"context"
	"fmt"
	"time"

	"goshort/internal/storage"

	"github.com/redis/go-redis/v9"
)

type redisRateLimiter struct {
	client           *redis.Client
	requestsPerMin   int
	windowSize       time.Duration
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, requestsPerMin int) storage.RateLimiter {
	return &redisRateLimiter{
		client:         client,
		requestsPerMin: requestsPerMin,
		windowSize:     time.Minute,
	}
}

func (r *redisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// Sanitize key
	key = sanitizeKey(fmt.Sprintf("ratelimit:%s", key))

	// Use sliding window algorithm
	now := time.Now().Unix()
	windowStart := now - int64(r.windowSize.Seconds())

	pipe := r.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current requests
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	// Set expiration
	pipe.Expire(ctx, key, r.windowSize+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	count := countCmd.Val()

	// Check if under limit
	return count < int64(r.requestsPerMin), nil
}

func (r *redisRateLimiter) Reset(ctx context.Context, key string) error {
	key = sanitizeKey(fmt.Sprintf("ratelimit:%s", key))

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	return nil
}

func (r *redisRateLimiter) GetRemaining(ctx context.Context, key string) (int64, error) {
	key = sanitizeKey(fmt.Sprintf("ratelimit:%s", key))

	now := time.Now().Unix()
	windowStart := now - int64(r.windowSize.Seconds())

	// Count requests in current window
	count, err := r.client.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get remaining requests: %w", err)
	}

	remaining := int64(r.requestsPerMin) - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

