package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// BackupToolCommand demonstrates a practical backup tool with layered configuration
type BackupToolCommand struct {
	*cmds.CommandDescription
}

// Configuration layers for backup tool
type BackupSettings struct {
	SourceDir       string   `glazed.parameter:"source-dir"`
	BackupDir       string   `glazed.parameter:"backup-dir"`
	IncludePatterns []string `glazed.parameter:"include-patterns"`
	ExcludePatterns []string `glazed.parameter:"exclude-patterns"`
	DryRun          bool     `glazed.parameter:"dry-run"`
	Recursive       bool     `glazed.parameter:"recursive"`
}

type CompressionSettings struct {
	Enabled      bool   `glazed.parameter:"compression-enabled"`
	Format       string `glazed.parameter:"format"`
	Level        int    `glazed.parameter:"level"`
	KeepOriginal bool   `glazed.parameter:"keep-original"`
}

type RetentionSettings struct {
	MaxAge         int  `glazed.parameter:"max-age-days"`
	MaxBackups     int  `glazed.parameter:"max-backups"`
	CleanupEnabled bool `glazed.parameter:"cleanup-enabled"`
}

type NotificationSettings struct {
	Enabled    bool   `glazed.parameter:"notification-enabled"`
	EmailTo    string `glazed.parameter:"email-to"`
	WebhookURL string `glazed.parameter:"webhook-url"`
	OnError    bool   `glazed.parameter:"on-error"`
	OnSuccess  bool   `glazed.parameter:"on-success"`
}

var _ cmds.GlazeCommand = &BackupToolCommand{}

func (c *BackupToolCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse all configuration layers
	backupSettings := &BackupSettings{}
	if err := parsedLayers.InitializeStruct("backup", backupSettings); err != nil {
		return err
	}

	compressionSettings := &CompressionSettings{}
	if err := parsedLayers.InitializeStruct("compression", compressionSettings); err != nil {
		return err
	}

	retentionSettings := &RetentionSettings{}
	if err := parsedLayers.InitializeStruct("retention", retentionSettings); err != nil {
		return err
	}

	notificationSettings := &NotificationSettings{}
	if err := parsedLayers.InitializeStruct("notification", notificationSettings); err != nil {
		return err
	}

	// Validate configuration
	if err := c.validateSettings(backupSettings, compressionSettings, retentionSettings); err != nil {
		return err
	}

	// Display configuration summary
	fmt.Printf("📁 Backup: %s → %s\n", backupSettings.SourceDir, backupSettings.BackupDir)
	if backupSettings.DryRun {
		fmt.Println("🧪 DRY RUN MODE - No files will be modified")
	}

	// Find files to backup
	files, err := c.findFilesToBackup(backupSettings)
	if err != nil {
		return err
	}

	// Process each file
	startTime := time.Now()
	var totalSize int64
	processed := 0

	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			continue
		}

		relativePath := strings.TrimPrefix(file, backupSettings.SourceDir)
		relativePath = strings.TrimPrefix(relativePath, "/")

		backupPath := filepath.Join(backupSettings.BackupDir, relativePath)

		// Simulate backup operation
		if !backupSettings.DryRun {
			// In a real implementation, you would copy the file here
			fmt.Printf("Backing up: %s\n", relativePath)
		}

		totalSize += fileInfo.Size()
		processed++

		// Create row for structured output
		row := types.NewRow(
			types.MRP("file", relativePath),
			types.MRP("source_path", file),
			types.MRP("backup_path", backupPath),
			types.MRP("size_bytes", fileInfo.Size()),
			types.MRP("modified_time", fileInfo.ModTime().Format(time.RFC3339)),
			types.MRP("backup_time", time.Now().Format(time.RFC3339)),
			types.MRP("compressed", compressionSettings.Enabled),
			types.MRP("compression_format", compressionSettings.Format),
			types.MRP("dry_run", backupSettings.DryRun),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	duration := time.Since(startTime)

	// Create summary row
	summaryRow := types.NewRow(
		types.MRP("operation", "backup_summary"),
		types.MRP("files_processed", processed),
		types.MRP("total_size_bytes", totalSize),
		types.MRP("total_size_mb", float64(totalSize)/(1024*1024)),
		types.MRP("duration_seconds", duration.Seconds()),
		types.MRP("files_per_second", float64(processed)/duration.Seconds()),
		types.MRP("compression_enabled", compressionSettings.Enabled),
		types.MRP("notification_enabled", notificationSettings.Enabled),
		types.MRP("cleanup_enabled", retentionSettings.CleanupEnabled),
	)

	if err := gp.AddRow(ctx, summaryRow); err != nil {
		return err
	}

	fmt.Printf("✅ Backup completed: %d files, %.2f MB in %.2fs\n",
		processed, float64(totalSize)/(1024*1024), duration.Seconds())

	return nil
}

func (c *BackupToolCommand) validateSettings(backup *BackupSettings, compression *CompressionSettings, retention *RetentionSettings) error {
	// Validate source directory exists
	if _, err := os.Stat(backup.SourceDir); os.IsNotExist(err) {
		return errors.Errorf("source directory does not exist: %s", backup.SourceDir)
	}

	// Validate backup directory can be created
	if err := os.MkdirAll(backup.BackupDir, 0755); err != nil {
		return errors.Errorf("cannot create backup directory: %s", err)
	}

	// Validate compression settings
	if compression.Enabled {
		validFormats := []string{"gzip", "bzip2", "xz", "zip"}
		found := false
		for _, format := range validFormats {
			if compression.Format == format {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("invalid compression format: %s (valid: %s)",
				compression.Format, strings.Join(validFormats, ", "))
		}

		if compression.Level < 1 || compression.Level > 9 {
			return errors.Errorf("compression level must be between 1 and 9, got: %d", compression.Level)
		}
	}

	// Validate retention settings
	if retention.MaxAge < 0 {
		return errors.Errorf("retention max-age-days cannot be negative: %d", retention.MaxAge)
	}

	if retention.MaxBackups < 0 {
		return errors.Errorf("retention max-backups cannot be negative: %d", retention.MaxBackups)
	}

	return nil
}

func (c *BackupToolCommand) findFilesToBackup(settings *BackupSettings) ([]string, error) {
	var files []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !settings.Recursive && path != settings.SourceDir {
				return filepath.SkipDir
			}
			return nil
		}

		// Check include patterns
		if len(settings.IncludePatterns) > 0 {
			matched := false
			for _, pattern := range settings.IncludePatterns {
				if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		// Check exclude patterns
		for _, pattern := range settings.ExcludePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
		}

		files = append(files, path)
		return nil
	}

	err := filepath.Walk(settings.SourceDir, walkFn)
	return files, err
}

func NewBackupToolCommand() (*BackupToolCommand, error) {
	// Create backup configuration layer
	backupLayer, err := layers.NewParameterLayer(
		"backup",
		"Backup Configuration",
		layers.WithDescription("Core backup settings"),
	)
	if err != nil {
		return nil, err
	}

	backupLayer.AddFlags(
		parameters.NewParameterDefinition(
			"source-dir",
			parameters.ParameterTypeString,
			parameters.WithHelp("Source directory to backup"),
			parameters.WithRequired(true),
		),
		parameters.NewParameterDefinition(
			"backup-dir",
			parameters.ParameterTypeString,
			parameters.WithHelp("Destination backup directory"),
			parameters.WithRequired(true),
		),
		parameters.NewParameterDefinition(
			"include-patterns",
			parameters.ParameterTypeStringList,
			parameters.WithHelp("File patterns to include (glob patterns)"),
			parameters.WithDefault([]string{}),
		),
		parameters.NewParameterDefinition(
			"exclude-patterns",
			parameters.ParameterTypeStringList,
			parameters.WithHelp("File patterns to exclude (glob patterns)"),
			parameters.WithDefault([]string{"*.tmp", "*.log", ".DS_Store"}),
		),
		parameters.NewParameterDefinition(
			"dry-run",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Show what would be backed up without doing it"),
			parameters.WithDefault(false),
		),
		parameters.NewParameterDefinition(
			"recursive",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Backup directories recursively"),
			parameters.WithDefault(true),
		),
	)

	// Create compression configuration layer
	compressionLayer, err := layers.NewParameterLayer(
		"compression",
		"Compression Configuration",
		layers.WithDescription("File compression settings"),
	)
	if err != nil {
		return nil, err
	}

	compressionLayer.AddFlags(
		parameters.NewParameterDefinition(
			"compression-enabled",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Enable compression of backup files"),
			parameters.WithDefault(false),
		),
		parameters.NewParameterDefinition(
			"format",
			parameters.ParameterTypeChoice,
			parameters.WithHelp("Compression format"),
			parameters.WithChoices("gzip", "bzip2", "xz", "zip"),
			parameters.WithDefault("gzip"),
		),
		parameters.NewParameterDefinition(
			"level",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Compression level (1-9, higher = better compression)"),
			parameters.WithDefault(6),
		),
		parameters.NewParameterDefinition(
			"keep-original",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Keep original files after compression"),
			parameters.WithDefault(true),
		),
	)

	// Create retention configuration layer
	retentionLayer, err := layers.NewParameterLayer(
		"retention",
		"Retention Configuration",
		layers.WithDescription("Backup retention and cleanup settings"),
	)
	if err != nil {
		return nil, err
	}

	retentionLayer.AddFlags(
		parameters.NewParameterDefinition(
			"max-age-days",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Maximum age of backups to keep (days, 0 = no limit)"),
			parameters.WithDefault(0),
		),
		parameters.NewParameterDefinition(
			"max-backups",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Maximum number of backups to keep (0 = no limit)"),
			parameters.WithDefault(0),
		),
		parameters.NewParameterDefinition(
			"cleanup-enabled",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Enable automatic cleanup of old backups"),
			parameters.WithDefault(false),
		),
	)

	// Create notification configuration layer
	notificationLayer, err := layers.NewParameterLayer(
		"notification",
		"Notification Configuration",
		layers.WithDescription("Backup completion notifications"),
	)
	if err != nil {
		return nil, err
	}

	notificationLayer.AddFlags(
		parameters.NewParameterDefinition(
			"notification-enabled",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Enable backup notifications"),
			parameters.WithDefault(false),
		),
		parameters.NewParameterDefinition(
			"email-to",
			parameters.ParameterTypeString,
			parameters.WithHelp("Email address for notifications"),
			parameters.WithDefault(""),
		),
		parameters.NewParameterDefinition(
			"webhook-url",
			parameters.ParameterTypeString,
			parameters.WithHelp("Webhook URL for notifications"),
			parameters.WithDefault(""),
		),
		parameters.NewParameterDefinition(
			"on-error",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Send notifications on errors"),
			parameters.WithDefault(true),
		),
		parameters.NewParameterDefinition(
			"on-success",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Send notifications on successful completion"),
			parameters.WithDefault(false),
		),
	)

	// Create glazed layer for structured output
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"backup",
		cmds.WithShort("Comprehensive backup tool with layered configuration"),
		cmds.WithLong(`
A comprehensive backup tool demonstrating practical use of parameter layers.

Configuration is organized into logical groups:
- Backup: Core backup settings (source, destination, patterns)
- Compression: File compression options
- Retention: Backup retention and cleanup policies
- Notification: Success/failure notifications
- Glazed: Structured output formatting

Examples:
  backup --source-dir ./documents --backup-dir ./backups
  backup --source-dir ./data --backup-dir ./backups --compression-enabled --format gzip
  backup --source-dir ./logs --backup-dir ./backups --include-patterns "*.log,*.txt"
  backup --source-dir ./web --backup-dir ./backups --dry-run --output table
		`),

		// Add all our custom layers
		cmds.WithLayersList(
			backupLayer,
			compressionLayer,
			retentionLayer,
			notificationLayer,
			glazedLayer,
		),
	)

	return &BackupToolCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "backup-tool",
		Short: "Comprehensive backup tool with layered configuration",
		Long: `
A practical backup tool demonstrating parameter layer organization
and configuration management for production-ready applications.
		`,
	}

	backupCmd, err := NewBackupToolCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(backupCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
