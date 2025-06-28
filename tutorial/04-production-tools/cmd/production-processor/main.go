package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tutorial-04/internal/config"
	apperrors "tutorial-04/internal/errors"
	"tutorial-04/internal/logging"
	"tutorial-04/internal/metrics"
	"tutorial-04/internal/processor"

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
	ResultFile   string `glazed.parameter:"result-file"`
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

	// Validate required parameters for file processing
	if settings.InputFile == "" {
		return apperrors.New(apperrors.ErrCodeInvalidInput, "input-file is required for file processing")
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
				parameters.WithRequired(false),
			),
			parameters.NewParameterDefinition(
				"parser",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log parser to use"),
				parameters.WithChoices("apache", "json", "simple"),
				parameters.WithDefault("apache"),
			),
			parameters.NewParameterDefinition(
				"result-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Result output file (optional)"),
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
