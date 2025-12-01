package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"goshort/internal/domain"
	"goshort/internal/service"

	"go.uber.org/zap"
)

type ShortenHandler struct {
	service service.URLShortener
	logger  *zap.SugaredLogger
	baseURL string
}

func NewShortenHandler(service service.URLShortener, logger *zap.SugaredLogger, baseURL string) *ShortenHandler {
	return &ShortenHandler{
		service: service,
		logger:  logger,
		baseURL: baseURL,
	}
}

type ShortenRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
}

type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}

func (h *ShortenHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest

	// Decode request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnw("invalid request body", "error", err)
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if strings.TrimSpace(req.URL) == "" {
		respondError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Get client IP
	clientIP := getClientIP(r)

	// Get user agent
	userAgent := r.UserAgent()

	// Call service
	url, err := h.service.ShortenURL(r.Context(), req.URL, req.CustomCode, clientIP, userAgent)
	if err != nil {
		h.handleServiceError(w, err, clientIP)
		return
	}

	// Build response
	response := ShortenResponse{
		ShortCode:   url.ShortCode,
		ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, url.ShortCode),
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	respondJSON(w, response, http.StatusCreated)
}

func (h *ShortenHandler) GetURLDetails(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/api/v1/urls/")

	url, err := h.service.GetURLDetails(r.Context(), shortCode)
	if err != nil {
		h.handleServiceError(w, err, "")
		return
	}

	response := map[string]interface{}{
		"short_code":   url.ShortCode,
		"original_url": url.OriginalURL,
		"created_at":   url.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"click_count":  url.ClickCount,
		"is_active":    url.IsActive,
	}

	respondJSON(w, response, http.StatusOK)
}

func (h *ShortenHandler) handleServiceError(w http.ResponseWriter, err error, clientIP string) {
	switch err {
	case domain.ErrURLNotFound:
		respondError(w, "URL not found", http.StatusNotFound)
	case domain.ErrURLExpired:
		respondError(w, "URL has expired", http.StatusGone)
	case domain.ErrURLInactive:
		respondError(w, "URL is inactive", http.StatusGone)
	case domain.ErrDuplicateShortCode:
		respondError(w, "short code already exists", http.StatusConflict)
	case domain.ErrRateLimitExceeded:
		h.logger.Warnw("rate limit exceeded", "ip", clientIP)
		respondError(w, "rate limit exceeded, please try again later", http.StatusTooManyRequests)
	case domain.ErrInvalidShortCode, domain.ErrInvalidURL:
		respondError(w, err.Error(), http.StatusBadRequest)
	default:
		if strings.Contains(err.Error(), "validation failed") || 
		   strings.Contains(err.Error(), "not allowed") ||
		   strings.Contains(err.Error(), "blocked") {
			h.logger.Warnw("validation error", "error", err, "ip", clientIP)
			respondError(w, "invalid or blocked URL", http.StatusBadRequest)
		} else {
			h.logger.Errorw("internal error", "error", err)
			respondError(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

