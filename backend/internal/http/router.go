package http

import (
	"net/http"
	"time"

	"goshort/internal/config"
	"goshort/internal/http/handlers"
	"goshort/internal/http/middleware"
	"goshort/internal/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// NewRouter creates a new HTTP router with all routes and middleware
func NewRouter(cfg *config.Config, logger *zap.SugaredLogger, urlService service.URLShortener) http.Handler {
	r := chi.NewRouter()

	// Standard middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// Custom logging middleware
	r.Use(LoggerMiddleware(logger))

	// Timeout middleware
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Security middleware
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.SecureHeaders(cfg.Security.TrustedProxies))

	// Request size limiter
	r.Use(middleware.RequestSizeLimiter(cfg.Security.MaxRequestBodySize))

	// CORS configuration
	if cfg.Security.EnableCORS {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.Security.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	// Rate limiting
	if cfg.Security.RateLimitEnabled {
		r.Use(middleware.RateLimiter(cfg.Security.RateLimitRequestsPerMin, cfg.Security.RateLimitBurst))
	}

	// Initialize handlers
	baseURL := getBaseURL(cfg)
	shortenHandler := handlers.NewShortenHandler(urlService, logger, baseURL)
	redirectHandler := handlers.NewRedirectHandler(urlService, logger)
	healthHandler := handlers.NewHealthHandler(logger)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// No cache for API endpoints
		r.Use(middleware.NoCache)

		// Health check
		r.Get("/health", healthHandler.Health)
		r.Get("/ready", healthHandler.Ready)

		// URL shortening endpoints
		r.Post("/shorten", shortenHandler.ShortenURL)
		r.Get("/urls/{shortCode}", shortenHandler.GetURLDetails)
	})

	// Short URL redirect (root level)
	r.Get("/{shortCode}", redirectHandler.Redirect)

	return r
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				logger.Infow("request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"duration_ms", time.Since(start).Milliseconds(),
					"request_id", chimiddleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// getBaseURL constructs the base URL for short links
func getBaseURL(cfg *config.Config) string {
	// In production, this should come from environment variable
	if cfg.Server.Environment == "production" {
		return "https://yourdomain.com"
	}
	return "http://localhost:8080"
}

