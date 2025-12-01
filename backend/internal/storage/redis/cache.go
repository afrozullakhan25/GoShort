package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"goshort/internal/storage"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache repository
func NewRedisCache(client *redis.Client) storage.CacheRepository {
	return &redisCache{client: client}
}

// Connect creates a new Redis client
func Connect(host string, port int, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	// Sanitize key
	key = sanitizeKey(key)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get from cache: %w", err)
	}

	return val, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value string, expiration int) error {
	// Sanitize key and value
	key = sanitizeKey(key)
	value = sanitizeValue(value)

	// Validate expiration (max 30 days)
	if expiration < 0 || expiration > 2592000 {
		expiration = 3600 // Default 1 hour
	}

	err := c.client.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	key = sanitizeKey(key)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	key = sanitizeKey(key)

	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists > 0, nil
}

func (c *redisCache) IncrementClickCount(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("clicks:%s", sanitizeKey(shortCode))

	err := c.client.Incr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	// Set expiration if key is new
	c.client.Expire(ctx, key, 24*time.Hour)

	return nil
}

func (c *redisCache) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	key := fmt.Sprintf("clicks:%s", sanitizeKey(shortCode))

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get click count: %w", err)
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse click count: %w", err)
	}

	return count, nil
}

// sanitizeKey removes potentially dangerous characters from cache keys
func sanitizeKey(key string) string {
	// Remove null bytes and control characters
	sanitized := ""
	for _, r := range key {
		if r >= 32 && r < 127 {
			sanitized += string(r)
		}
	}

	// Limit length to prevent memory issues
	if len(sanitized) > 250 {
		sanitized = sanitized[:250]
	}

	return sanitized
}

// sanitizeValue sanitizes cache values
func sanitizeValue(value string) string {
	// Limit value size to prevent memory issues (1MB max)
	if len(value) > 1048576 {
		value = value[:1048576]
	}
	return value
}

