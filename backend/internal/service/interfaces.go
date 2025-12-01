package service

import (
	"context"
	"goshort/internal/domain"
)

// URLShortener defines the interface for URL shortening service
type URLShortener interface {
	// ShortenURL creates a short URL from original URL
	ShortenURL(ctx context.Context, originalURL, customCode, clientIP, userAgent string) (*domain.URL, error)
	
	// GetOriginalURL retrieves original URL by short code
	GetOriginalURL(ctx context.Context, shortCode string) (*domain.URL, error)
	
	// GetURLDetails retrieves URL details with stats
	GetURLDetails(ctx context.Context, shortCode string) (*domain.URL, error)
	
	// DeleteURL soft deletes a URL
	DeleteURL(ctx context.Context, id string) error
	
	// ListURLs lists URLs with pagination
	ListURLs(ctx context.Context, limit, offset int) ([]*domain.URL, error)
}

