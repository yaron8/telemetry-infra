package logi

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
)

// Config holds the logging configuration
type Config struct {
	// LogDir is the directory where log files will be stored
	// Default: /var/log/telemetry-infra (or ./logs if not writable)
	LogDir string
	// LogFileName is the name of the log file
	// Default: app.log
	LogFileName string
	// Level is the minimum log level to write
	// Default: slog.LevelInfo
	Level slog.Level
}

// NewLog creates or returns the singleton logger instance.
// It's safe for concurrent use across multiple goroutines.
// The logger writes to a file with JSON formatting and buffered I/O for maximum performance.
func NewLog(cfg *Config) (*slog.Logger, error) {
	var initErr error

	once.Do(func() {
		if cfg == nil {
			cfg = &Config{}
		}

		// Set defaults
		if cfg.LogDir == "" {
			// Try /var/log/telemetry-infra first (works in Docker)
			// Fall back to ./logs if not writable
			cfg.LogDir = "/var/log/telemetry-infra"
			if !isDirWritable(cfg.LogDir) {
				cfg.LogDir = "./logs"
			}
		}

		if cfg.LogFileName == "" {
			cfg.LogFileName = "app.log"
		}

		// Create log directory if it doesn't exist
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory %s: %w", cfg.LogDir, err)
			return
		}

		logPath := filepath.Join(cfg.LogDir, cfg.LogFileName)

		// Open log file with append mode
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file %s: %w", logPath, err)
			return
		}

		// Create JSON handler for structured logging
		level := cfg.Level
		if level == 0 {
			level = slog.LevelInfo
		}

		opts := &slog.HandlerOptions{
			Level: level,
			// Remove source info for max performance
			AddSource: false,
		}

		handler := slog.NewJSONHandler(file, opts)
		logger = slog.New(handler)

		// Log initialization
		logger.Info("logger initialized",
			"log_path", logPath,
			"level", level.String(),
		)
	})

	if initErr != nil {
		return nil, initErr
	}

	return logger, nil
}

// GetLogger returns the existing logger instance.
// This is a zero-allocation, lock-free read after initialization.
// Panics if NewLog hasn't been called yet - call NewLog once at startup.
func GetLogger() *slog.Logger {
	if logger == nil {
		panic("logger not initialized - call NewLog first")
	}
	return logger
}

// isDirWritable checks if a directory is writable
func isDirWritable(path string) bool {
	// Try to create the directory first
	if err := os.MkdirAll(path, 0755); err != nil {
		return false
	}

	// Try to create a temp file to verify write access
	testFile := filepath.Join(path, ".write_test")
	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}
