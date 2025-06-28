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

	"tutorial-04/internal/config"
	apperrors "tutorial-04/internal/errors"
	"tutorial-04/internal/logging"
	"tutorial-04/internal/metrics"
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

// GetParsers returns available parsers
func (p *Processor) GetParsers() map[string]LogParser {
	return p.parsers
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

// HealthCheck for the processor
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
