package handlers

import (
	"net/http"

	"goshort/internal/domain"
	"goshort/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type RedirectHandler struct {
	service service.URLShortener
	logger  *zap.SugaredLogger
}

func NewRedirectHandler(service service.URLShortener, logger *zap.SugaredLogger) *RedirectHandler {
	return &RedirectHandler{
		service: service,
		logger:  logger,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	// Validate short code format
	if err := domain.ValidateShortCode(shortCode); err != nil {
		h.logger.Warnw("invalid short code", "code", shortCode, "error", err)
		http.Error(w, "Invalid short code", http.StatusBadRequest)
		return
	}

	// Get original URL
	url, err := h.service.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		h.handleRedirectError(w, err, shortCode)
		return
	}

	// Log redirect
	h.logger.Infow("redirecting",
		"short_code", shortCode,
		"original_url", url.OriginalURL,
		"ip", getClientIP(r),
	)

	// Perform redirect with 301 (permanent)
	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}

func (h *RedirectHandler) handleRedirectError(w http.ResponseWriter, err error, shortCode string) {
	switch err {
	case domain.ErrURLNotFound:
		http.Error(w, "Short URL not found", http.StatusNotFound)
	case domain.ErrURLExpired:
		http.Error(w, "Short URL has expired", http.StatusGone)
	case domain.ErrURLInactive:
		http.Error(w, "Short URL is inactive", http.StatusGone)
	default:
		h.logger.Errorw("redirect error", "error", err, "short_code", shortCode)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

