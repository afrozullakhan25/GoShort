package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Environment  string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host        string
	Port        int
	Password    string
	DB          int
	MaxRetries  int
	PoolSize    int
	MinIdleConns int
}

type SecurityConfig struct {
	// SSRF Protection
	AllowedDomains          []string
	UseAllowlist            bool
	AllowedPorts            []int
	MaxRedirects            int
	TimeoutSeconds          int
	DisableIPLiterals       bool
	DNSRevalidationCount    int
	DNSRevalidationDelayMs  int
	
	// Rate Limiting
	RateLimitEnabled        bool
	RateLimitRequestsPerMin int
	RateLimitBurst          int
	
	// General Security
	EnableCORS              bool
	AllowedOrigins          []string
	MaxRequestBodySize      int64
	TrustedProxies          []string
	
	// Short Code Generation
	ShortCodeLength         int
	ShortCodeAlphabet       string
}

type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", "10s"),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", "10s"),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", "120s"),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "goshort"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvAsInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
		},
		Security: SecurityConfig{
			AllowedDomains:          getEnvAsSlice("SECURITY_ALLOWED_DOMAINS", ""),
			UseAllowlist:            getEnvAsBool("SECURITY_USE_ALLOWLIST", true),
			AllowedPorts:            getEnvAsIntSlice("SECURITY_ALLOWED_PORTS", "80,443"),
			MaxRedirects:            getEnvAsInt("SECURITY_MAX_REDIRECTS", 0),
			TimeoutSeconds:          getEnvAsInt("SECURITY_TIMEOUT_SECONDS", 10),
			DisableIPLiterals:       getEnvAsBool("SECURITY_DISABLE_IP_LITERALS", true),
			DNSRevalidationCount:    getEnvAsInt("SECURITY_DNS_REVALIDATION_COUNT", 2),
			DNSRevalidationDelayMs:  getEnvAsInt("SECURITY_DNS_REVALIDATION_DELAY_MS", 100),
			RateLimitEnabled:        getEnvAsBool("SECURITY_RATE_LIMIT_ENABLED", true),
			RateLimitRequestsPerMin: getEnvAsInt("SECURITY_RATE_LIMIT_RPM", 60),
			RateLimitBurst:          getEnvAsInt("SECURITY_RATE_LIMIT_BURST", 10),
			EnableCORS:              getEnvAsBool("SECURITY_ENABLE_CORS", false),
			AllowedOrigins:          getEnvAsSlice("SECURITY_ALLOWED_ORIGINS", ""),
			MaxRequestBodySize:      getEnvAsInt64("SECURITY_MAX_REQUEST_BODY_SIZE", 1048576),
			TrustedProxies:          getEnvAsSlice("SECURITY_TRUSTED_PROXIES", ""),
			ShortCodeLength:         getEnvAsInt("SHORT_CODE_LENGTH", 8),
			ShortCodeAlphabet:       getEnv("SHORT_CODE_ALPHABET", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", "stdout"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Database validation
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	// Security validation
	if c.Security.UseAllowlist && len(c.Security.AllowedDomains) == 0 {
		return fmt.Errorf("allowlist enabled but no domains specified")
	}
	if len(c.Security.AllowedPorts) == 0 {
		return fmt.Errorf("no allowed ports specified")
	}
	if c.Security.ShortCodeLength < 4 || c.Security.ShortCodeLength > 20 {
		return fmt.Errorf("invalid short code length: %d", c.Security.ShortCodeLength)
	}

	// Logging validation
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}
	if !validLogLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}

func getEnvAsSlice(key string, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	if value == "" {
		return []string{}
	}
	
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getEnvAsIntSlice(key string, defaultValue string) []int {
	strSlice := getEnvAsSlice(key, defaultValue)
	result := make([]int, 0, len(strSlice))
	
	for _, str := range strSlice {
		if intVal, err := strconv.Atoi(str); err == nil {
			result = append(result, intVal)
		}
	}
	
	return result
}

