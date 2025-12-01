package logging

import (
	"strings"

	"goshort/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new structured logger
func NewLogger(cfg *config.Config) *zap.SugaredLogger {
	var zapConfig zap.Config

	// Determine log level
	level := parseLogLevel(cfg.Logging.Level)

	// Configure based on environment
	if cfg.Server.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(level)
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(level)
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set output format
	if strings.ToLower(cfg.Logging.Format) == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// Set output path
	if cfg.Logging.OutputPath != "" && cfg.Logging.OutputPath != "stdout" {
		zapConfig.OutputPaths = []string{cfg.Logging.OutputPath}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
	}

	// Build logger
	logger, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}

	return logger.Sugar()
}

// parseLogLevel converts string log level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

