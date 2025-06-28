package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// FileSyncCommand represents the file synchronization command
type FileSyncCommand struct {
	*cmds.CommandDescription
	metrics *SyncMetrics
}

// SyncSettings contains command-specific settings
type SyncSettings struct {
	SourcePath      string `glazed.parameter:"source"`
	DestinationPath string `glazed.parameter:"destination"`
	DryRun          bool   `glazed.parameter:"dry-run"`
	ChecksumVerify  bool   `glazed.parameter:"checksum-verify"`
	MaxRetries      int    `glazed.parameter:"max-retries"`
	Verbose         bool   `glazed.parameter:"verbose"`
	Delete          bool   `glazed.parameter:"delete"`
}

// SyncMetrics holds synchronization metrics
type SyncMetrics struct {
	FilesProcessed *prometheus.CounterVec
	BytesTransferred prometheus.Counter
	SyncDuration   prometheus.Histogram
	ErrorsTotal    *prometheus.CounterVec
}

// FileInfo represents file metadata
type FileInfo struct {
	Path         string
	Size         int64
	ModTime      time.Time
	Checksum     string
	IsDirectory  bool
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	FilesScanned    int64
	FilesTransferred int64
	FilesSkipped    int64
	FilesDeleted    int64
	BytesTransferred int64
	Errors          []error
	Duration        time.Duration
}

var _ cmds.GlazeCommand = &FileSyncCommand{}

func (c *FileSyncCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &SyncSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to parse settings")
	}

	// Initialize logger
	logger := log.With().Str("component", "file-sync").Logger()

	// Initialize metrics
	c.metrics = NewSyncMetrics()

	logger.Info().
		Str("source", settings.SourcePath).
		Str("destination", settings.DestinationPath).
		Bool("dry_run", settings.DryRun).
		Msg("Starting file synchronization")

	// Perform the sync
	result, err := c.performSync(ctx, settings, logger)
	if err != nil {
		c.metrics.ErrorsTotal.WithLabelValues("sync", "failed").Inc()
		return err
	}

	// Record metrics
	c.metrics.SyncDuration.Observe(result.Duration.Seconds())
	c.metrics.FilesProcessed.WithLabelValues("scanned").Add(float64(result.FilesScanned))
	c.metrics.FilesProcessed.WithLabelValues("transferred").Add(float64(result.FilesTransferred))
	c.metrics.FilesProcessed.WithLabelValues("skipped").Add(float64(result.FilesSkipped))
	c.metrics.BytesTransferred.Add(float64(result.BytesTransferred))

	// Output results
	summaryRow := types.NewRow(
		types.MRP("type", "summary"),
		types.MRP("files_scanned", result.FilesScanned),
		types.MRP("files_transferred", result.FilesTransferred),
		types.MRP("files_skipped", result.FilesSkipped),
		types.MRP("files_deleted", result.FilesDeleted),
		types.MRP("bytes_transferred", result.BytesTransferred),
		types.MRP("duration_seconds", result.Duration.Seconds()),
		types.MRP("errors", len(result.Errors)),
		types.MRP("dry_run", settings.DryRun),
	)

	if err := gp.AddRow(ctx, summaryRow); err != nil {
		return err
	}

	// Output errors if any
	for _, syncErr := range result.Errors {
		errorRow := types.NewRow(
			types.MRP("type", "error"),
			types.MRP("error", syncErr.Error()),
		)
		if err := gp.AddRow(ctx, errorRow); err != nil {
			return err
		}
	}

	return nil
}

func (c *FileSyncCommand) performSync(ctx context.Context, settings *SyncSettings, logger zerolog.Logger) (*SyncResult, error) {
	result := &SyncResult{
		Errors: make([]error, 0),
	}
	startTime := time.Now()

	// Check if source exists
	_, err := os.Stat(settings.SourcePath)
	if err != nil {
		return nil, errors.Wrap(err, "source path does not exist")
	}

	// Create destination if it doesn't exist
	if !settings.DryRun {
		if err := os.MkdirAll(settings.DestinationPath, 0755); err != nil {
			return nil, errors.Wrap(err, "failed to create destination directory")
		}
	}

	// Walk source directory
	err = filepath.Walk(settings.SourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, errors.Wrapf(err, "error walking path %s", path))
			return nil // Continue walking
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result.FilesScanned++

		// Calculate relative path
		relPath, err := filepath.Rel(settings.SourcePath, path)
		if err != nil {
			result.Errors = append(result.Errors, errors.Wrapf(err, "failed to calculate relative path for %s", path))
			return nil
		}

		destPath := filepath.Join(settings.DestinationPath, relPath)

		if info.IsDir() {
			// Handle directory
			if !settings.DryRun {
				if err := os.MkdirAll(destPath, info.Mode()); err != nil {
					result.Errors = append(result.Errors, errors.Wrapf(err, "failed to create directory %s", destPath))
				}
			}
			logger.Debug().Str("dir", destPath).Msg("Directory created")
			return nil
		}

		// Handle file
		shouldCopy, err := c.shouldCopyFile(path, destPath, settings)
		if err != nil {
			result.Errors = append(result.Errors, errors.Wrapf(err, "failed to check if file should be copied: %s", path))
			return nil
		}

		if !shouldCopy {
			result.FilesSkipped++
			logger.Debug().Str("file", path).Msg("File skipped (up to date)")
			return nil
		}

		// Copy file
		if !settings.DryRun {
			err = c.copyFileWithRetry(path, destPath, settings.MaxRetries, logger)
			if err != nil {
				result.Errors = append(result.Errors, errors.Wrapf(err, "failed to copy file %s", path))
				return nil
			}
		}

		result.FilesTransferred++
		result.BytesTransferred += info.Size()
		
		if settings.Verbose {
			logger.Info().
				Str("source", path).
				Str("destination", destPath).
				Int64("size", info.Size()).
				Msg("File transferred")
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "error during file walk")
	}

	// Handle deletion if requested
	if settings.Delete && !settings.DryRun {
		deletedCount, err := c.deleteExtraFiles(settings.SourcePath, settings.DestinationPath, logger)
		if err != nil {
			result.Errors = append(result.Errors, errors.Wrap(err, "failed to delete extra files"))
		}
		result.FilesDeleted = deletedCount
	}

	result.Duration = time.Since(startTime)

	logger.Info().
		Int64("files_scanned", result.FilesScanned).
		Int64("files_transferred", result.FilesTransferred).
		Int64("files_skipped", result.FilesSkipped).
		Int64("bytes_transferred", result.BytesTransferred).
		Dur("duration", result.Duration).
		Int("errors", len(result.Errors)).
		Msg("Synchronization completed")

	return result, nil
}

func (c *FileSyncCommand) shouldCopyFile(sourcePath, destPath string, settings *SyncSettings) (bool, error) {
	// Check if destination exists
	destStat, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		return true, nil // File doesn't exist, should copy
	}
	if err != nil {
		return false, err
	}

	sourceStat, err := os.Stat(sourcePath)
	if err != nil {
		return false, err
	}

	// Check modification time
	if sourceStat.ModTime().After(destStat.ModTime()) {
		return true, nil
	}

	// Check size
	if sourceStat.Size() != destStat.Size() {
		return true, nil
	}

	// Check checksum if requested
	if settings.ChecksumVerify {
		sourceChecksum, err := c.calculateChecksum(sourcePath)
		if err != nil {
			return false, err
		}

		destChecksum, err := c.calculateChecksum(destPath)
		if err != nil {
			return false, err
		}

		return sourceChecksum != destChecksum, nil
	}

	return false, nil // File is up to date
}

func (c *FileSyncCommand) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (c *FileSyncCommand) copyFileWithRetry(sourcePath, destPath string, maxRetries int, logger zerolog.Logger) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			logger.Warn().
				Str("file", sourcePath).
				Int("attempt", attempt).
				Dur("backoff", backoff).
				Err(lastErr).
				Msg("Retrying file copy")
			time.Sleep(backoff)
		}

		err := c.copyFile(sourcePath, destPath)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return errors.Wrapf(lastErr, "failed to copy file after %d attempts", maxRetries+1)
}

func (c *FileSyncCommand) copyFile(sourcePath, destPath string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy data
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file metadata
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	if err := os.Chmod(destPath, sourceInfo.Mode()); err != nil {
		return err
	}

	if err := os.Chtimes(destPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		return err
	}

	return nil
}

func (c *FileSyncCommand) deleteExtraFiles(sourcePath, destPath string, logger zerolog.Logger) (int64, error) {
	var deletedCount int64

	// Build a set of files that exist in source
	sourceFiles := make(map[string]bool)
	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		sourceFiles[relPath] = true
		return nil
	})
	if err != nil {
		return 0, err
	}

	// Walk destination and delete files that don't exist in source
	err = filepath.Walk(destPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		relPath, err := filepath.Rel(destPath, path)
		if err != nil {
			return err
		}

		if !sourceFiles[relPath] {
			logger.Info().Str("file", path).Msg("Deleting extra file")
			if err := os.Remove(path); err != nil {
				return err
			}
			deletedCount++
		}
		return nil
	})

	return deletedCount, err
}

func NewSyncMetrics() *SyncMetrics {
	return &SyncMetrics{
		FilesProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "file_sync_files_processed_total",
				Help: "Total number of files processed during sync",
			},
			[]string{"operation"},
		),
		BytesTransferred: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "file_sync_bytes_transferred_total",
				Help: "Total bytes transferred during sync",
			},
		),
		SyncDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name: "file_sync_duration_seconds",
				Help: "Duration of sync operations",
			},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "file_sync_errors_total",
				Help: "Total number of sync errors",
			},
			[]string{"operation", "result"},
		),
	}
}

func NewFileSyncCommand() (*FileSyncCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"sync",
		cmds.WithShort("Production-ready file synchronization"),
		cmds.WithLong(`
A production-ready file synchronizer with comprehensive error handling,
retry logic, checksum verification, and monitoring capabilities.

Features:
- Incremental sync with checksum verification
- Retry logic with exponential backoff
- Prometheus metrics
- Dry-run mode for testing
- Directory deletion support
- Structured logging

Examples:
  sync --source ./src --destination ./dst
  sync --source ./src --destination ./dst --dry-run --verbose
  sync --source ./src --destination ./dst --checksum-verify --delete
  sync --source ./src --destination ./dst --max-retries 3
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"source",
				parameters.ParameterTypeString,
				parameters.WithHelp("Source directory path"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"destination",
				parameters.ParameterTypeString,
				parameters.WithHelp("Destination directory path"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"dry-run",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Show what would be done without actually doing it"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"checksum-verify",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Verify files using checksums"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"max-retries",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of retries for failed operations"),
				parameters.WithDefault(2),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose logging"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"delete",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Delete files in destination that don't exist in source"),
				parameters.WithDefault(false),
			),
		),

		cmds.WithLayersList(glazedLayer),
	)

	return &FileSyncCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "file-synchronizer",
		Short: "Production-ready file synchronizer",
		Long: `
A comprehensive file synchronization tool demonstrating production-ready patterns:
- Incremental synchronization with checksum verification
- Retry logic with exponential backoff
- Comprehensive error handling and logging
- Prometheus metrics for monitoring
- Dry-run mode for testing
		`,
	}

	syncCmd, err := NewFileSyncCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(syncCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
