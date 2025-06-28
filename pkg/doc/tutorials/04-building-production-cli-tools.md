---
Title: Building Production CLI Tools
Slug: building-production-cli-tools
Short: Learn to create robust, production-ready CLI applications with proper error handling, testing, and deployment
Topics:
- production
- testing
- error-handling
- deployment
- tutorial
Commands:
Flags:
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Building Production CLI Tools

This tutorial teaches you how to create robust, production-ready CLI applications using glazed. You'll learn about error handling, logging, testing strategies, configuration management, and deployment best practices.

## Learning Objectives

- Implement comprehensive error handling and logging
- Create testable command structures
- Build configuration management systems
- Add health checks and monitoring
- Deploy and distribute CLI applications
- Handle performance and security considerations

## Prerequisites

- Completed Tutorials 1, 2, and 3
- Understanding of Go testing frameworks
- Basic knowledge of CI/CD concepts
- Familiarity with deployment strategies

## Tutorial Overview

We'll build a production-ready "log-processor" tool that demonstrates:
- Robust error handling and recovery
- Comprehensive logging and monitoring
- Unit and integration testing
- Configuration management
- Performance optimization
- Security best practices
- Deployment strategies

## Setting Up

```bash
mkdir glazed-tutorial-04
cd glazed-tutorial-04
go mod init glazed-tutorial-04
go get github.com/go-go-golems/glazed
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/rs/zerolog
go get github.com/prometheus/client_golang/prometheus
```

## Part 1: Error Handling and Logging

Let's start with a robust foundation for error handling and structured logging.

Create `internal/errors/errors.go`:

```go
package errors

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrCodeFileNotFound    ErrorCode = "FILE_NOT_FOUND"
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrCodeProcessingError ErrorCode = "PROCESSING_ERROR"
	ErrCodeNetworkError    ErrorCode = "NETWORK_ERROR"
	ErrCodeConfigError     ErrorCode = "CONFIG_ERROR"
	ErrCodeInternalError   ErrorCode = "INTERNAL_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
	Stack   string                 `json:"stack,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code ErrorCode, message string, details ...map[string]interface{}) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
		Stack:   getStack(),
	}
	
	if len(details) > 0 {
		err.Details = details[0]
	}
	
	return err
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string, details ...map[string]interface{}) *AppError {
	appErr := &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
		Stack:   getStack(),
	}
	
	if len(details) > 0 {
		appErr.Details = details[0]
	}
	
	return appErr
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	
	var stack string
	for {
		frame, more := frames.Next()
		stack += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return stack
}
```

Create `internal/logging/logger.go`:

```go
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
```

Create `internal/metrics/metrics.go`:

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all application metrics
type Metrics struct {
	CommandExecutions *prometheus.CounterVec
	CommandDuration   *prometheus.HistogramVec
	ErrorsTotal       *prometheus.CounterVec
	FilesProcessed    prometheus.Counter
	BytesProcessed    prometheus.Counter
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		CommandExecutions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "logprocessor_command_executions_total",
				Help: "Total number of command executions",
			},
			[]string{"command", "status"},
		),
		CommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "logprocessor_command_duration_seconds",
				Help:    "Duration of command executions",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"command"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "logprocessor_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "code"},
		),
		FilesProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "logprocessor_files_processed_total",
				Help: "Total number of files processed",
			},
		),
		BytesProcessed: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "logprocessor_bytes_processed_total",
				Help: "Total number of bytes processed",
			},
		),
	}
}

// Default metrics instance
var defaultMetrics *Metrics

// InitDefaultMetrics initializes the default metrics
func InitDefaultMetrics() {
	defaultMetrics = NewMetrics()
}

// GetDefaultMetrics returns the default metrics instance
func GetDefaultMetrics() *Metrics {
	if defaultMetrics == nil {
		InitDefaultMetrics()
	}
	return defaultMetrics
}
```

## Part 2: Configuration Management

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"glazed-tutorial-04/internal/logging"
	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	App     AppConfig     `yaml:"app" json:"app"`
	Logging logging.Config `yaml:"logging" json:"logging"`
	Server  ServerConfig  `yaml:"server" json:"server"`
	Storage StorageConfig `yaml:"storage" json:"storage"`
}

// AppConfig contains application-specific settings
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
	Debug       bool   `yaml:"debug" json:"debug"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	Host           string `yaml:"host" json:"host"`
	Port           int    `yaml:"port" json:"port"`
	MetricsEnabled bool   `yaml:"metrics_enabled" json:"metrics_enabled"`
	MetricsPort    int    `yaml:"metrics_port" json:"metrics_port"`
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Type      string            `yaml:"type" json:"type"`
	Path      string            `yaml:"path" json:"path"`
	Options   map[string]string `yaml:"options" json:"options"`
	Retention int               `yaml:"retention_days" json:"retention_days"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "log-processor",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       false,
		},
		Logging: logging.Config{
			Level:      "info",
			Format:     "console",
			Output:     "stderr",
			AddCaller:  false,
			TimeFormat: "2006-01-02T15:04:05Z07:00",
		},
		Server: ServerConfig{
			Host:           "localhost",
			Port:           8080,
			MetricsEnabled: true,
			MetricsPort:    9090,
		},
		Storage: StorageConfig{
			Type:      "local",
			Path:      "./data",
			Options:   map[string]string{},
			Retention: 30,
		},
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Load from file if it exists
	if configPath != "" && fileExists(configPath) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if env := os.Getenv("LOG_PROCESSOR_ENV"); env != "" {
		config.App.Environment = env
	}
	if debug := os.Getenv("LOG_PROCESSOR_DEBUG"); debug == "true" {
		config.App.Debug = true
	}
	if logLevel := os.Getenv("LOG_PROCESSOR_LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name is required")
	}
	
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Storage.Type == "" {
		return fmt.Errorf("storage type is required")
	}
	
	if c.Storage.Retention < 0 {
		return fmt.Errorf("storage retention must be non-negative")
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```

## Part 3: Testable Command Structure

Create `internal/processor/processor.go`:

```go
package processor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"glazed-tutorial-04/internal/config"
	apperrors "glazed-tutorial-04/internal/errors"
	"glazed-tutorial-04/internal/logging"
	"glazed-tutorial-04/internal/metrics"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Source    string
	Message   string
	Fields    map[string]interface{}
	Raw       string
}

// Processor handles log processing
type Processor struct {
	config  *config.Config
	logger  *logging.Logger
	metrics *metrics.Metrics
	parsers map[string]LogParser
}

// LogParser interface for different log formats
type LogParser interface {
	Parse(line string) (*LogEntry, error)
	Name() string
}

// ProcessorOption for configuring the processor
type ProcessorOption func(*Processor)

// NewProcessor creates a new log processor
func NewProcessor(cfg *config.Config, opts ...ProcessorOption) *Processor {
	p := &Processor{
		config:  cfg,
		logger:  logging.GetDefaultLogger(),
		metrics: metrics.GetDefaultMetrics(),
		parsers: make(map[string]LogParser),
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	// Register default parsers
	p.RegisterParser(&ApacheLogParser{})
	p.RegisterParser(&JSONLogParser{})
	p.RegisterParser(&SimpleLogParser{})

	return p
}

// WithLogger sets a custom logger
func WithLogger(logger *logging.Logger) ProcessorOption {
	return func(p *Processor) {
		p.logger = logger
	}
}

// WithMetrics sets custom metrics
func WithMetrics(metrics *metrics.Metrics) ProcessorOption {
	return func(p *Processor) {
		p.metrics = metrics
	}
}

// RegisterParser registers a log parser
func (p *Processor) RegisterParser(parser LogParser) {
	p.parsers[parser.Name()] = parser
}

// ProcessFile processes a single log file
func (p *Processor) ProcessFile(ctx context.Context, filePath string, parserName string) (*ProcessResult, error) {
	startTime := time.Now()
	
	p.logger.Info().
		Str("file", filePath).
		Str("parser", parserName).
		Msg("Starting file processing")

	parser, exists := p.parsers[parserName]
	if !exists {
		err := apperrors.New(
			apperrors.ErrCodeInvalidInput,
			"unknown parser",
			map[string]interface{}{"parser": parserName},
		)
		p.metrics.ErrorsTotal.WithLabelValues("parser", string(apperrors.ErrCodeInvalidInput)).Inc()
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		err = apperrors.Wrap(
			err,
			apperrors.ErrCodeFileNotFound,
			"failed to open file",
			map[string]interface{}{"file": filePath},
		)
		p.metrics.ErrorsTotal.WithLabelValues("file", string(apperrors.ErrCodeFileNotFound)).Inc()
		return nil, err
	}
	defer file.Close()

	result, err := p.processReader(ctx, file, parser)
	if err != nil {
		return nil, err
	}

	// Update metrics
	duration := time.Since(startTime)
	p.metrics.CommandDuration.WithLabelValues("process-file").Observe(duration.Seconds())
	p.metrics.FilesProcessed.Inc()
	p.metrics.BytesProcessed.Add(float64(result.BytesProcessed))

	p.logger.Info().
		Str("file", filePath).
		Int64("lines_processed", result.LinesProcessed).
		Int64("lines_failed", result.LinesFailed).
		Dur("duration", duration).
		Msg("File processing completed")

	return result, nil
}

// ProcessResult contains processing statistics
type ProcessResult struct {
	LinesProcessed int64
	LinesFailed    int64
	BytesProcessed int64
	Entries        []*LogEntry
	Errors         []error
}

func (p *Processor) processReader(ctx context.Context, reader io.Reader, parser LogParser) (*ProcessResult, error) {
	result := &ProcessResult{
		Entries: make([]*LogEntry, 0),
		Errors:  make([]error, 0),
	}

	scanner := bufio.NewScanner(reader)
	lineNumber := int64(0)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		lineNumber++
		line := scanner.Text()
		result.BytesProcessed += int64(len(line))

		entry, err := parser.Parse(line)
		if err != nil {
			result.LinesFailed++
			result.Errors = append(result.Errors, apperrors.Wrap(
				err,
				apperrors.ErrCodeProcessingError,
				"failed to parse line",
				map[string]interface{}{
					"line_number": lineNumber,
					"line":        line,
				},
			))
			
			p.logger.Debug().
				Err(err).
				Int64("line_number", lineNumber).
				Str("line", line).
				Msg("Failed to parse line")
			continue
		}

		result.LinesProcessed++
		result.Entries = append(result.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return result, apperrors.Wrap(
			err,
			apperrors.ErrCodeProcessingError,
			"error reading file",
		)
	}

	return result, nil
}

// Health check for the processor
func (p *Processor) HealthCheck() error {
	// Check if storage is accessible
	if p.config.Storage.Type == "local" {
		if _, err := os.Stat(p.config.Storage.Path); os.IsNotExist(err) {
			return apperrors.New(
				apperrors.ErrCodeConfigError,
				"storage path does not exist",
				map[string]interface{}{"path": p.config.Storage.Path},
			)
		}
	}

	// Check parser availability
	if len(p.parsers) == 0 {
		return apperrors.New(
			apperrors.ErrCodeInternalError,
			"no parsers registered",
		)
	}

	return nil
}

// Apache Common Log Format parser
type ApacheLogParser struct{}

func (p *ApacheLogParser) Name() string {
	return "apache"
}

func (p *ApacheLogParser) Parse(line string) (*LogEntry, error) {
	// Apache Common Log Format: IP - - [timestamp] "method url protocol" status size
	re := regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "([^"]*)" (\d+) (\S+)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 6 {
		return nil, fmt.Errorf("invalid apache log format")
	}

	timestamp, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	status, _ := strconv.Atoi(matches[4])
	size := matches[5]
	if size == "-" {
		size = "0"
	}

	return &LogEntry{
		Timestamp: timestamp,
		Level:     "info",
		Source:    "apache",
		Message:   matches[3],
		Fields: map[string]interface{}{
			"ip":     matches[1],
			"status": status,
			"size":   size,
		},
		Raw: line,
	}, nil
}

// JSON log parser
type JSONLogParser struct{}

func (p *JSONLogParser) Name() string {
	return "json"
}

func (p *JSONLogParser) Parse(line string) (*LogEntry, error) {
	// Simple JSON parser - in production, use a proper JSON library
	if !strings.HasPrefix(line, "{") {
		return nil, fmt.Errorf("not a JSON log line")
	}

	// This is a simplified implementation
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Source:    "json",
		Message:   line,
		Fields:    make(map[string]interface{}),
		Raw:       line,
	}

	return entry, nil
}

// Simple log parser
type SimpleLogParser struct{}

func (p *SimpleLogParser) Name() string {
	return "simple"
}

func (p *SimpleLogParser) Parse(line string) (*LogEntry, error) {
	// Simple format: TIMESTAMP LEVEL MESSAGE
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid simple log format")
	}

	timestamp, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		// Try alternative format
		timestamp, err = time.Parse("2006-01-02 15:04:05", parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	return &LogEntry{
		Timestamp: timestamp,
		Level:     strings.ToLower(parts[1]),
		Source:    "simple",
		Message:   parts[2],
		Fields:    make(map[string]interface{}),
		Raw:       line,
	}, nil
}
```

## Part 4: Production Command Implementation

Create `cmd/production-processor/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"glazed-tutorial-04/internal/config"
	apperrors "glazed-tutorial-04/internal/errors"
	"glazed-tutorial-04/internal/logging"
	"glazed-tutorial-04/internal/metrics"
	"glazed-tutorial-04/internal/processor"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

// ProductionProcessorCommand represents the main processing command
type ProductionProcessorCommand struct {
	*cmds.CommandDescription
	config    *config.Config
	processor *processor.Processor
	logger    *logging.Logger
	metrics   *metrics.Metrics
}

// ProcessorSettings contains command-specific settings
type ProcessorSettings struct {
	InputFile    string `glazed.parameter:"input-file"`
	Parser       string `glazed.parameter:"parser"`
	OutputFile   string `glazed.parameter:"output-file"`
	ConfigFile   string `glazed.parameter:"config-file"`
	Timeout      int    `glazed.parameter:"timeout"`
	MaxErrors    int    `glazed.parameter:"max-errors"`
	HealthCheck  bool   `glazed.parameter:"health-check"`
	StartMetrics bool   `glazed.parameter:"start-metrics"`
	Verbose      bool   `glazed.parameter:"verbose"`
}

var _ cmds.GlazeCommand = &ProductionProcessorCommand{}

func (c *ProductionProcessorCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ProcessorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeInvalidInput, "failed to parse settings")
	}

	// Load configuration
	if err := c.loadConfiguration(settings.ConfigFile); err != nil {
		return err
	}

	// Initialize logger
	logging.InitDefaultLogger(c.config.Logging)
	c.logger = logging.GetDefaultLogger()

	// Initialize metrics
	metrics.InitDefaultMetrics()
	c.metrics = metrics.GetDefaultMetrics()

	// Start metrics server if requested
	if settings.StartMetrics && c.config.Server.MetricsEnabled {
		go c.startMetricsServer()
	}

	// Health check
	if settings.HealthCheck {
		return c.performHealthCheck(ctx, gp)
	}

	// Process file
	return c.processFile(ctx, gp, settings)
}

func (c *ProductionProcessorCommand) loadConfiguration(configFile string) error {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeConfigError, "failed to load configuration")
	}

	if err := cfg.Validate(); err != nil {
		return apperrors.Wrap(err, apperrors.ErrCodeConfigError, "invalid configuration")
	}

	c.config = cfg
	return nil
}

func (c *ProductionProcessorCommand) startMetricsServer() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := c.processor.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(fmt.Sprintf("Health check failed: %v", err)))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Ready check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.config.Server.MetricsPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	c.logger.Info().
		Int("port", c.config.Server.MetricsPort).
		Msg("Starting metrics server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		c.logger.Error().Err(err).Msg("Metrics server error")
	}
}

func (c *ProductionProcessorCommand) performHealthCheck(ctx context.Context, gp middlewares.Processor) error {
	c.logger.Info().Msg("Performing health check")

	// Initialize processor
	c.processor = processor.NewProcessor(
		c.config,
		processor.WithLogger(c.logger),
		processor.WithMetrics(c.metrics),
	)

	// Perform health check
	if err := c.processor.HealthCheck(); err != nil {
		c.metrics.ErrorsTotal.WithLabelValues("health", string(apperrors.ErrCodeInternalError)).Inc()
		
		row := types.NewRow(
			types.MRP("status", "unhealthy"),
			types.MRP("error", err.Error()),
			types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
		
		return err
	}

	row := types.NewRow(
		types.MRP("status", "healthy"),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		types.MRP("config_valid", true),
		types.MRP("parsers_available", len(c.processor.GetParsers())),
	)

	return gp.AddRow(ctx, row)
}

func (c *ProductionProcessorCommand) processFile(ctx context.Context, gp middlewares.Processor, settings *ProcessorSettings) error {
	startTime := time.Now()
	
	// Create timeout context
	if settings.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(settings.Timeout)*time.Second)
		defer cancel()
	}

	// Initialize processor
	c.processor = processor.NewProcessor(
		c.config,
		processor.WithLogger(c.logger),
		processor.WithMetrics(c.metrics),
	)

	c.logger.Info().
		Str("file", settings.InputFile).
		Str("parser", settings.Parser).
		Msg("Starting file processing")

	// Process the file
	result, err := c.processor.ProcessFile(ctx, settings.InputFile, settings.Parser)
	if err != nil {
		c.metrics.CommandExecutions.WithLabelValues("process", "error").Inc()
		return err
	}

	c.metrics.CommandExecutions.WithLabelValues("process", "success").Inc()

	// Check error threshold
	if settings.MaxErrors > 0 && int(result.LinesFailed) > settings.MaxErrors {
		err := apperrors.New(
			apperrors.ErrCodeProcessingError,
			"too many processing errors",
			map[string]interface{}{
				"errors":     result.LinesFailed,
				"max_errors": settings.MaxErrors,
			},
		)
		c.metrics.ErrorsTotal.WithLabelValues("threshold", string(apperrors.ErrCodeProcessingError)).Inc()
		return err
	}

	// Output summary
	duration := time.Since(startTime)
	
	summaryRow := types.NewRow(
		types.MRP("type", "summary"),
		types.MRP("file", settings.InputFile),
		types.MRP("parser", settings.Parser),
		types.MRP("lines_processed", result.LinesProcessed),
		types.MRP("lines_failed", result.LinesFailed),
		types.MRP("bytes_processed", result.BytesProcessed),
		types.MRP("duration_seconds", duration.Seconds()),
		types.MRP("lines_per_second", float64(result.LinesProcessed)/duration.Seconds()),
		types.MRP("success_rate", float64(result.LinesProcessed)/float64(result.LinesProcessed+result.LinesFailed)*100),
	)

	if err := gp.AddRow(ctx, summaryRow); err != nil {
		return err
	}

	// Output detailed entries if verbose
	if settings.Verbose {
		for i, entry := range result.Entries {
			if i >= 100 { // Limit output in verbose mode
				break
			}
			
			entryRow := types.NewRow(
				types.MRP("type", "entry"),
				types.MRP("timestamp", entry.Timestamp.Format(time.RFC3339)),
				types.MRP("level", entry.Level),
				types.MRP("source", entry.Source),
				types.MRP("message", entry.Message),
			)

			// Add fields
			for k, v := range entry.Fields {
				entryRow.Set(k, v)
			}

			if err := gp.AddRow(ctx, entryRow); err != nil {
				return err
			}
		}
	}

	// Output errors if any
	if len(result.Errors) > 0 {
		for i, err := range result.Errors {
			if i >= 10 { // Limit error output
				break
			}
			
			var errorCode string
			var details map[string]interface{}
			
			if appErr, ok := err.(*apperrors.AppError); ok {
				errorCode = string(appErr.Code)
				details = appErr.Details
			}

			errorRow := types.NewRow(
				types.MRP("type", "error"),
				types.MRP("error", err.Error()),
				types.MRP("code", errorCode),
			)

			for k, v := range details {
				errorRow.Set(k, v)
			}

			if err := gp.AddRow(ctx, errorRow); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetParsers returns available parsers (for health check)
func (p *processor.Processor) GetParsers() map[string]processor.LogParser {
	// This would need to be added to the processor package
	// For now, return empty map
	return make(map[string]processor.LogParser)
}

func NewProductionProcessorCommand() (*ProductionProcessorCommand, error) {
	// Create glazed layer
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"process",
		cmds.WithShort("Production-ready log processing"),
		cmds.WithLong(`
A production-ready log processor with comprehensive error handling,
monitoring, health checks, and configuration management.

Features:
- Multiple log format parsers (Apache, JSON, Simple)
- Prometheus metrics
- Health check endpoints
- Configurable timeouts and error thresholds
- Structured logging
- Graceful error handling

Examples:
  process --input-file access.log --parser apache
  process --input-file app.log --parser json --verbose --output json
  process --health-check --start-metrics
  process --config-file config.yaml --input-file logs/*.log --timeout 300
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"input-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Input log file to process"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"parser",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log parser to use"),
				parameters.WithChoices("apache", "json", "simple"),
				parameters.WithDefault("apache"),
			),
			parameters.NewParameterDefinition(
				"output-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Output file (optional)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"config-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Configuration file path"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"timeout",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Processing timeout in seconds (0 = no timeout)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"max-errors",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum allowed errors before stopping (0 = no limit)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"health-check",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Perform health check and exit"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"start-metrics",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Start metrics server"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include detailed log entries in output"),
				parameters.WithDefault(false),
			),
		),

		cmds.WithLayersList(glazedLayer),
	)

	return &ProductionProcessorCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, gracefully shutting down...")
		cancel()
	}()

	rootCmd := &cobra.Command{
		Use:   "production-processor",
		Short: "Production-ready log processor",
		Long: `
A comprehensive log processing tool demonstrating production-ready patterns:
- Robust error handling and recovery
- Structured logging and metrics
- Health checks and monitoring
- Configuration management
- Graceful shutdown
		`,
	}

	processorCmd, err := NewProductionProcessorCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(processorCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

## Part 5: Testing Strategy

Create `internal/processor/processor_test.go`:

```go
package processor

import (
	"context"
	"strings"
	"testing"
	"time"

	"glazed-tutorial-04/internal/config"
	"glazed-tutorial-04/internal/logging"
	"glazed-tutorial-04/internal/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessor_ProcessReader(t *testing.T) {
	cfg := config.DefaultConfig()
	processor := NewProcessor(cfg)

	tests := []struct {
		name           string
		input          string
		parser         string
		expectedLines  int64
		expectedErrors int64
	}{
		{
			name:           "valid apache logs",
			input:          `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
			parser:         "apache",
			expectedLines:  1,
			expectedErrors: 0,
		},
		{
			name:           "invalid apache logs",
			input:          `invalid log line`,
			parser:         "apache",
			expectedLines:  0,
			expectedErrors: 1,
		},
		{
			name: "mixed valid and invalid",
			input: `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
invalid line
192.168.1.1 - - [10/Oct/2000:13:55:37 -0700] "POST /form HTTP/1.0" 404 1234`,
			parser:         "apache",
			expectedLines:  2,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := processor.parsers[tt.parser]
			require.NotNil(t, parser, "Parser not found: %s", tt.parser)

			result, err := processor.processReader(context.Background(), reader, parser)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedLines, result.LinesProcessed)
			assert.Equal(t, tt.expectedErrors, result.LinesFailed)
		})
	}
}

func TestApacheLogParser_Parse(t *testing.T) {
	parser := &ApacheLogParser{}

	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedIP  string
		expectedStatus int
	}{
		{
			name:           "valid log line",
			input:          `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
			expectError:    false,
			expectedIP:     "127.0.0.1",
			expectedStatus: 200,
		},
		{
			name:        "invalid log line",
			input:       `invalid log format`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parser.Parse(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, entry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
				assert.Equal(t, tt.expectedIP, entry.Fields["ip"])
				assert.Equal(t, tt.expectedStatus, entry.Fields["status"])
			}
		})
	}
}

func TestProcessor_HealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      config.DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid storage path",
			config: &config.Config{
				Storage: config.StorageConfig{
					Type: "local",
					Path: "/nonexistent/path",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(tt.config)
			err := processor.HealthCheck()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkApacheLogParser_Parse(b *testing.B) {
	parser := &ApacheLogParser{}
	logLine := `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(logLine)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessor_ProcessReader(b *testing.B) {
	cfg := config.DefaultConfig()
	processor := NewProcessor(cfg)
	parser := &ApacheLogParser{}

	// Create a larger input
	logLines := make([]string, 1000)
	for i := range logLines {
		logLines[i] = `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`
	}
	input := strings.Join(logLines, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(input)
		_, err := processor.processReader(context.Background(), reader, parser)
		if err != nil {
			b.Fatal(err)
		}
	}
}
```

Create `cmd/production-processor/main_test.go`:

```go
package main

import (
	"context"
	"os"
	"testing"
	"time"

	"glazed-tutorial-04/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionProcessorCommand_Integration(t *testing.T) {
	// Create a temporary log file
	logContent := `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
192.168.1.1 - - [10/Oct/2000:13:55:37 -0700] "POST /form HTTP/1.0" 404 1234
10.0.0.1 - - [10/Oct/2000:13:55:38 -0700] "GET /index.html HTTP/1.0" 200 5432
`

	tmpFile, err := os.CreateTemp("", "test-access-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(logContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test the command
	cmd, err := NewProductionProcessorCommand()
	require.NoError(t, err)

	// This would require more setup to test the full command
	// For now, just test that it creates successfully
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.CommandDescription)
}

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	configContent := `
app:
  name: test-processor
  version: 1.0.0
  environment: test

logging:
  level: debug
  format: console

server:
  host: localhost
  port: 8080
  metrics_enabled: true
  metrics_port: 9090

storage:
  type: local
  path: ./test-data
  retention_days: 7
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load the config
	cfg, err := config.LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "test-processor", cfg.App.Name)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, 9090, cfg.Server.MetricsPort)
	assert.Equal(t, 7, cfg.Storage.Retention)
}
```

### Test Your Production Command

```bash
# Create sample log file
echo '127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
192.168.1.1 - - [10/Oct/2000:13:55:37 -0700] "POST /form HTTP/1.0" 404 1234' > sample.log

# Health check
go run cmd/production-processor/main.go process --health-check

# Process with metrics
go run cmd/production-processor/main.go process --input-file sample.log --parser apache --start-metrics --verbose

# Run tests
go test ./internal/processor/
go test ./cmd/production-processor/

# Run benchmarks
go test -bench=. ./internal/processor/
```

## Key Concepts Learned

### Production Readiness
- **Error Handling**: Structured errors with codes and context
- **Logging**: Structured logging with different levels and formats
- **Metrics**: Prometheus metrics for monitoring
- **Health Checks**: Endpoint for monitoring system health

### Testing Strategy
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **Benchmarks**: Performance testing and optimization
- **Mocks**: Test doubles for external dependencies

### Configuration Management
- **Defaults**: Sensible default values
- **File Loading**: YAML/JSON configuration files
- **Environment Variables**: Runtime configuration override
- **Validation**: Ensure configuration is valid and complete

### Deployment Considerations
- **Graceful Shutdown**: Handle shutdown signals properly
- **Timeouts**: Prevent hanging operations
- **Resource Management**: Clean up resources properly
- **Monitoring**: Provide observability into application behavior

## Summary

This tutorial covered building production-ready CLI tools with:
- Comprehensive error handling and recovery
- Structured logging and metrics collection
- Robust configuration management
- Testing strategies and performance optimization
- Health checks and monitoring capabilities
- Graceful shutdown and resource management

These patterns and practices will help you build CLI tools that are reliable, maintainable, and suitable for production deployment.

## Exercise

Build a "production-ready file synchronizer" with these requirements:

1. **Core Functionality**:
   - Sync files between local and remote locations
   - Support multiple protocols (S3, FTP, local filesystem)
   - Handle incremental syncs with checksums
   - Retry failed transfers with exponential backoff

2. **Production Features**:
   - Comprehensive error handling with custom error types
   - Structured logging with different output formats
   - Prometheus metrics for monitoring sync operations
   - Health check endpoints for service monitoring
   - Configuration management with environment variable overrides

3. **Testing**:
   - Unit tests for all core components
   - Integration tests with mock external services
   - Benchmark tests for performance validation
   - Property-based tests for edge case discovery

4. **Deployment**:
   - Docker container support
   - Kubernetes deployment manifests
   - CI/CD pipeline configuration
   - Monitoring and alerting setup

Test your implementation thoroughly and ensure it handles all error conditions gracefully!
