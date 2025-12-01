package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// Validation errors
	ErrInvalidURL      = errors.New("invalid URL format")
	ErrURLTooLong      = errors.New("URL exceeds maximum length")
	ErrEmptyURL        = errors.New("URL cannot be empty")
	ErrInvalidShortCode = errors.New("invalid short code format")
)

const (
	MaxURLLength       = 2048
	MaxShortCodeLength = 50
	MinShortCodeLength = 4
)

// ShortCode regex: alphanumeric only, prevent path traversal
var shortCodeRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type URL struct {
	ID           string    `json:"id" db:"id"`
	OriginalURL  string    `json:"original_url" db:"original_url"`
	ShortCode    string    `json:"short_code" db:"short_code"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	ClickCount   int64     `json:"click_count" db:"click_count"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedByIP  string    `json:"-" db:"created_by_ip"`
	UserAgent    string    `json:"-" db:"user_agent"`
}

// NewURL creates a new URL with validation
func NewURL(originalURL, shortCode, createdByIP, userAgent string) (*URL, error) {
	// Validate original URL
	if err := ValidateOriginalURL(originalURL); err != nil {
		return nil, err
	}

	// Validate short code
	if err := ValidateShortCode(shortCode); err != nil {
		return nil, err
	}

	// Sanitize inputs
	sanitizedURL := SanitizeURL(originalURL)
	sanitizedCode := SanitizeShortCode(shortCode)
	sanitizedIP := SanitizeIP(createdByIP)
	sanitizedUA := SanitizeUserAgent(userAgent)

	return &URL{
		OriginalURL: sanitizedURL,
		ShortCode:   sanitizedCode,
		CreatedAt:   time.Now().UTC(),
		IsActive:    true,
		CreatedByIP: sanitizedIP,
		UserAgent:   sanitizedUA,
		ClickCount:  0,
	}, nil
}

// ValidateOriginalURL validates the original URL
func ValidateOriginalURL(url string) error {
	// Check empty
	url = strings.TrimSpace(url)
	if url == "" {
		return ErrEmptyURL
	}

	// Check length
	if len(url) > MaxURLLength {
		return ErrURLTooLong
	}

	// Check valid UTF-8
	if !utf8.ValidString(url) {
		return errors.New("URL contains invalid UTF-8 characters")
	}

	// Check for null bytes
	if strings.Contains(url, "\x00") {
		return errors.New("URL contains null bytes")
	}

	// Check for control characters
	for _, r := range url {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return errors.New("URL contains control characters")
		}
	}

	return nil
}

// ValidateShortCode validates the short code format
func ValidateShortCode(code string) error {
	code = strings.TrimSpace(code)
	
	if code == "" {
		return ErrInvalidShortCode
	}

	if len(code) < MinShortCodeLength || len(code) > MaxShortCodeLength {
		return errors.New("short code length must be between 4 and 50 characters")
	}

	// Only alphanumeric, dash, underscore (prevent path traversal)
	if !shortCodeRegex.MatchString(code) {
		return ErrInvalidShortCode
	}

	// Prevent reserved words and patterns
	reservedWords := []string{
		"admin", "api", "login", "logout", "register", 
		"health", "metrics", "static", "assets", "public",
		"..", ".", "~", "null", "undefined",
	}

	codeLower := strings.ToLower(code)
	for _, reserved := range reservedWords {
		if codeLower == reserved || strings.Contains(codeLower, reserved) {
			return errors.New("short code contains reserved words")
		}
	}

	return nil
}

// SanitizeURL removes dangerous characters from URL
func SanitizeURL(url string) string {
	// Remove null bytes
	url = strings.ReplaceAll(url, "\x00", "")
	
	// Remove control characters except tab, newline, carriage return
	var sanitized strings.Builder
	for _, r := range url {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			sanitized.WriteRune(r)
		}
	}
	
	// Trim whitespace
	return strings.TrimSpace(sanitized.String())
}

// SanitizeShortCode sanitizes short code
func SanitizeShortCode(code string) string {
	// Remove non-alphanumeric except dash and underscore
	var sanitized strings.Builder
	for _, r := range code {
		if (r >= 'a' && r <= 'z') || 
		   (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || 
		   r == '-' || r == '_' {
			sanitized.WriteRune(r)
		}
	}
	return sanitized.String()
}

// SanitizeIP sanitizes IP address
func SanitizeIP(ip string) string {
	// Extract IP from "IP:port" format
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	
	// Remove square brackets from IPv6
	ip = strings.TrimPrefix(ip, "[")
	ip = strings.TrimSuffix(ip, "]")
	
	// Limit length
	if len(ip) > 45 { // Max IPv6 length
		ip = ip[:45]
	}
	
	return strings.TrimSpace(ip)
}

// SanitizeUserAgent sanitizes user agent string
func SanitizeUserAgent(ua string) string {
	// Limit length
	if len(ua) > 500 {
		ua = ua[:500]
	}
	
	// Remove control characters
	var sanitized strings.Builder
	for _, r := range ua {
		if r >= 32 && r < 127 {
			sanitized.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(sanitized.String())
}

// IsExpired checks if URL has expired
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*u.ExpiresAt)
}

// IncrementClick increments the click count
func (u *URL) IncrementClick() {
	u.ClickCount++
}

