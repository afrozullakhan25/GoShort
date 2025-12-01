package storage

import (
	"context"
	"goshort/internal/domain"
)

// URLRepository defines methods for URL storage operations
type URLRepository interface {
	// Create creates a new URL record
	Create(ctx context.Context, url *domain.URL) error
	
	// GetByShortCode retrieves URL by short code
	GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error)
	
	// GetByID retrieves URL by ID
	GetByID(ctx context.Context, id string) (*domain.URL, error)
	
	// Update updates an existing URL record
	Update(ctx context.Context, url *domain.URL) error
	
	// Delete soft deletes a URL record
	Delete(ctx context.Context, id string) error
	
	// IncrementClickCount increments the click count for a URL
	IncrementClickCount(ctx context.Context, shortCode string) error
	
	// Exists checks if short code already exists
	Exists(ctx context.Context, shortCode string) (bool, error)
	
	// List retrieves URLs with pagination
	List(ctx context.Context, limit, offset int) ([]*domain.URL, error)
}

// CacheRepository defines methods for caching operations
type CacheRepository interface {
	// Get retrieves value from cache
	Get(ctx context.Context, key string) (string, error)
	
	// Set stores value in cache with expiration
	Set(ctx context.Context, key string, value string, expiration int) error
	
	// Delete removes value from cache
	Delete(ctx context.Context, key string) error
	
	// Exists checks if key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
	
	// IncrementClickCount increments click count in cache
	IncrementClickCount(ctx context.Context, shortCode string) error
	
	// GetClickCount retrieves click count from cache
	GetClickCount(ctx context.Context, shortCode string) (int64, error)
}

// RateLimiter defines methods for rate limiting
type RateLimiter interface {
	// Allow checks if request is allowed based on rate limit
	Allow(ctx context.Context, key string) (bool, error)
	
	// Reset resets the rate limit for a key
	Reset(ctx context.Context, key string) error
	
	// GetRemaining returns remaining requests for a key
	GetRemaining(ctx context.Context, key string) (int64, error)
}

