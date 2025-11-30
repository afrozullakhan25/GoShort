package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	port := getEnv("PORT", "8080")
	r := gin.New()

	// -------------------------------------------------------------------
	// Middlewares
	// -------------------------------------------------------------------

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(securityHeaders())
	r.Use(requestTimeout(5 * time.Second))
	r.Use(rateLimiter(20, time.Minute)) // 20 requests per minute per IP

	// -------------------------------------------------------------------
	// Routes
	// -------------------------------------------------------------------

	r.GET("/", healthHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/shorten", shortenHandler)

	// -------------------------------------------------------------------
	// HTTP Server (Anti-Slowloris + Hardening)
	// -------------------------------------------------------------------

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second, // protects from Slowloris
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run server
	go func() {
		log.Printf("GoShort running on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Startup failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}

//
// ----------------------- HANDLERS -------------------------------------
//

func healthHandler(c *gin.Context) {
	hostname, _ := os.Hostname()
	c.JSON(http.StatusOK, gin.H{
		"service": "GoShort",
		"version": "v1.0.0",
		"status":  "healthy",
		"server":  hostname,
		"uptime":  time.Now().Format(time.RFC3339),
	})
}

func shortenHandler(c *gin.Context) {
	var req ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	shortURL := "https://goshort.ly/" + "xyz123" // TODO: real implementation

	c.JSON(http.StatusOK, gin.H{
		"original":  req.URL,
		"shortened": shortURL,
		"source":    "Mock DB",
	})
}

//
// ----------------------- MIDDLEWARES ----------------------------------
//

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// Request timeout middleware
func requestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Rate Limiter: X requests per duration per IP
func rateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	tokens := make(map[string]int)
	timestamps := make(map[string]time.Time)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// Reset window
		if t, ok := timestamps[ip]; !ok || now.Sub(t) > window {
			tokens[ip] = 0
			timestamps[ip] = now
		}

		// Check limit
		if tokens[ip] >= limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Try again later.",
			})
			c.Abort()
			return
		}

		// Consume token
		tokens[ip]++
		c.Next()
	}
}

//
// ----------------------- HELPERS --------------------------------------
//

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
