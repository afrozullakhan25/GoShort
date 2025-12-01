package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

		// HSTS (only for HTTPS)
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		// Remove server header
		w.Header().Del("Server")
		w.Header().Set("X-Powered-By", "")

		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimiter limits the size of request bodies
func RequestSizeLimiter(maxBytes int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders middleware prevents common attacks
func SecureHeaders(trustedProxies []string) func(next http.Handler) http.Handler {
	trustedMap := make(map[string]bool)
	for _, ip := range trustedProxies {
		trustedMap[strings.TrimSpace(ip)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate X-Forwarded headers only from trusted proxies
			clientIP := r.RemoteAddr
			if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
				clientIP = clientIP[:idx]
			}

			// If not from trusted proxy, remove forwarded headers
			if len(trustedMap) > 0 && !trustedMap[clientIP] {
				r.Header.Del("X-Forwarded-For")
				r.Header.Del("X-Forwarded-Host")
				r.Header.Del("X-Forwarded-Proto")
				r.Header.Del("X-Real-IP")
			}

			// Prevent HTTP header injection
			for key := range r.Header {
				for _, value := range r.Header[key] {
					if strings.ContainsAny(value, "\r\n") {
						http.Error(w, "Invalid header value", http.StatusBadRequest)
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NoCache prevents caching of sensitive endpoints
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

