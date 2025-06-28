package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog with application-specific functionality
type Logger struct {
	logger zerolog.Logger
}

// Config for logger setup
type Config struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	AddCaller  bool   `yaml:"add_caller" json:"add_caller"`
	TimeFormat string `yaml:"time_format" json:"time_format"`
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config Config) *Logger {
	// Set log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output zerolog.ConsoleWriter
	if config.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger
	var logger zerolog.Logger
	if config.Format == "console" {
		logger = log.Output(output)
	} else {
		logger = log.Logger
	}

	// Add caller info if requested
	if config.AddCaller {
		logger = logger.With().Caller().Logger()
	}

	return &Logger{logger: logger}
}

// Info logs an info message
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Error logs an error message
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Debug logs a debug message
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Warn logs a warning message
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// With adds fields to the logger
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

// Default logger instance
var defaultLogger *Logger

// InitDefaultLogger initializes the default logger
func InitDefaultLogger(config Config) {
	defaultLogger = NewLogger(config)
}

// GetDefaultLogger returns the default logger
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		InitDefaultLogger(Config{
			Level:  "info",
			Format: "console",
		})
	}
	return defaultLogger
}

// Convenience functions using default logger
func Info() *zerolog.Event {
	return GetDefaultLogger().Info()
}

func Error() *zerolog.Event {
	return GetDefaultLogger().Error()
}

func Debug() *zerolog.Event {
	return GetDefaultLogger().Debug()
}

func Warn() *zerolog.Event {
	return GetDefaultLogger().Warn()
}

func Fatal() *zerolog.Event {
	return GetDefaultLogger().Fatal()
}
