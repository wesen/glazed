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
	"github.com/spf13/cobra"
)

// FileProcessorCommand demonstrates layered parameter management
type FileProcessorCommand struct {
	*cmds.CommandDescription
}

// Settings structs for different layers
type DatabaseSettings struct {
	Host     string `glazed.parameter:"host"`
	Port     int    `glazed.parameter:"port"`
	Database string `glazed.parameter:"database"`
	Username string `glazed.parameter:"username"`
	Password string `glazed.parameter:"password"`
}

type ProcessingSettings struct {
	BatchSize    int      `glazed.parameter:"batch-size"`
	Workers      int      `glazed.parameter:"workers"`
	FileTypes    []string `glazed.parameter:"file-types"`
	ExcludeFiles []string `glazed.parameter:"exclude-files"`
	MaxFileSize  int      `glazed.parameter:"max-file-size"`
}

type OutputSettings struct {
	Format      string `glazed.parameter:"format"`
	Destination string `glazed.parameter:"destination"`
	Compress    bool   `glazed.parameter:"compress"`
}

// DefaultSettings for command-specific parameters
type DefaultSettings struct {
	InputDir  string `glazed.parameter:"input-dir"`
	Verbose   bool   `glazed.parameter:"verbose"`
	DryRun    bool   `glazed.parameter:"dry-run"`
	Recursive bool   `glazed.parameter:"recursive"`
}

var _ cmds.GlazeCommand = &FileProcessorCommand{}

func (c *FileProcessorCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from different layers
	defaultSettings := &DefaultSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, defaultSettings); err != nil {
		return err
	}

	dbSettings := &DatabaseSettings{}
	if err := parsedLayers.InitializeStruct("database", dbSettings); err != nil {
		return err
	}

	processingSettings := &ProcessingSettings{}
	if err := parsedLayers.InitializeStruct("processing", processingSettings); err != nil {
		return err
	}

	outputSettings := &OutputSettings{}
	if err := parsedLayers.InitializeStruct("output", outputSettings); err != nil {
		return err
	}

	if defaultSettings.Verbose {
		fmt.Printf("Processing directory: %s\n", defaultSettings.InputDir)
		fmt.Printf("Database: %s@%s:%d/%s\n",
			dbSettings.Username, dbSettings.Host, dbSettings.Port, dbSettings.Database)
		fmt.Printf("Workers: %d, Batch size: %d\n",
			processingSettings.Workers, processingSettings.BatchSize)
		fmt.Printf("Output: %s (%s)\n", outputSettings.Destination, outputSettings.Format)
	}

	// Simulate processing files
	files, err := c.findFiles(defaultSettings.InputDir, processingSettings, defaultSettings.Recursive)
	if err != nil {
		return err
	}

	for i, file := range files {
		if defaultSettings.DryRun {
			fmt.Printf("Would process: %s\n", file)
		}

		// Create row for each file
		row := types.NewRow(
			types.MRP("id", i+1),
			types.MRP("file", file),
			types.MRP("size", c.getFileSize(file)),
			types.MRP("type", filepath.Ext(file)),
			types.MRP("processed_at", time.Now().Format(time.RFC3339)),
			types.MRP("worker_id", (i%processingSettings.Workers)+1),
			types.MRP("batch", i/processingSettings.BatchSize+1),
			types.MRP("dry_run", defaultSettings.DryRun),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *FileProcessorCommand) findFiles(inputDir string, settings *ProcessingSettings, recursive bool) ([]string, error) {
	var files []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != inputDir {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file type filter
		ext := strings.ToLower(filepath.Ext(path))
		if len(settings.FileTypes) > 0 {
			found := false
			for _, allowedType := range settings.FileTypes {
				if ext == "."+strings.ToLower(allowedType) {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// Check exclude filter
		for _, excludePattern := range settings.ExcludeFiles {
			if matched, _ := filepath.Match(excludePattern, filepath.Base(path)); matched {
				return nil
			}
		}

		// Check file size
		if settings.MaxFileSize > 0 && info.Size() > int64(settings.MaxFileSize) {
			return nil
		}

		files = append(files, path)
		return nil
	}

	err := filepath.Walk(inputDir, walkFn)
	return files, err
}

func (c *FileProcessorCommand) getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func NewFileProcessorCommand() (*FileProcessorCommand, error) {
	// Create custom layers for different configuration groups

	// Database configuration layer
	dbLayer, err := layers.NewParameterLayer(
		"database",
		"Database Configuration",
		layers.WithDescription("Database connection settings"),
	)
	if err != nil {
		return nil, err
	}

	dbLayer.AddFlags(
		parameters.NewParameterDefinition(
			"host",
			parameters.ParameterTypeString,
			parameters.WithHelp("Database host"),
			parameters.WithDefault("localhost"),
		),
		parameters.NewParameterDefinition(
			"port",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Database port"),
			parameters.WithDefault(5432),
		),
		parameters.NewParameterDefinition(
			"database",
			parameters.ParameterTypeString,
			parameters.WithHelp("Database name"),
			parameters.WithDefault("fileprocessor"),
		),
		parameters.NewParameterDefinition(
			"username",
			parameters.ParameterTypeString,
			parameters.WithHelp("Database username"),
			parameters.WithDefault("admin"),
		),
		parameters.NewParameterDefinition(
			"password",
			parameters.ParameterTypeString,
			parameters.WithHelp("Database password"),
			parameters.WithDefault(""),
		),
	)

	// Processing configuration layer
	processingLayer, err := layers.NewParameterLayer(
		"processing",
		"Processing Configuration",
		layers.WithDescription("File processing settings"),
	)
	if err != nil {
		return nil, err
	}

	processingLayer.AddFlags(
		parameters.NewParameterDefinition(
			"batch-size",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Number of files to process in each batch"),
			parameters.WithDefault(10),
		),
		parameters.NewParameterDefinition(
			"workers",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Number of worker processes"),
			parameters.WithDefault(4),
		),
		parameters.NewParameterDefinition(
			"file-types",
			parameters.ParameterTypeStringList,
			parameters.WithHelp("File types to process (e.g., txt,json,csv)"),
			parameters.WithDefault([]string{}),
		),
		parameters.NewParameterDefinition(
			"exclude-files",
			parameters.ParameterTypeStringList,
			parameters.WithHelp("File patterns to exclude (glob patterns)"),
			parameters.WithDefault([]string{}),
		),
		parameters.NewParameterDefinition(
			"max-file-size",
			parameters.ParameterTypeInteger,
			parameters.WithHelp("Maximum file size in bytes (0 = no limit)"),
			parameters.WithDefault(0),
		),
	)

	// Output configuration layer
	outputLayer, err := layers.NewParameterLayer(
		"output",
		"Output Configuration",
		layers.WithDescription("Output formatting and destination settings"),
	)
	if err != nil {
		return nil, err
	}

	outputLayer.AddFlags(
		parameters.NewParameterDefinition(
			"format",
			parameters.ParameterTypeChoice,
			parameters.WithHelp("Output format"),
			parameters.WithChoices("json", "csv", "xml", "parquet"),
			parameters.WithDefault("json"),
		),
		parameters.NewParameterDefinition(
			"destination",
			parameters.ParameterTypeString,
			parameters.WithHelp("Output destination (file path or URL)"),
			parameters.WithDefault("./output"),
		),
		parameters.NewParameterDefinition(
			"compress",
			parameters.ParameterTypeBool,
			parameters.WithHelp("Compress output files"),
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
		"process",
		cmds.WithShort("Process files with configurable layers"),
		cmds.WithLong(`
A file processing command that demonstrates parameter layer organization.

Configuration is organized into logical groups:
- Default: Basic command options
- Database: Database connection settings  
- Processing: File processing configuration
- Output: Output format and destination
- Glazed: Structured output options

Examples:
  process --input-dir ./data --workers 8 --output json
  process --input-dir ./logs --file-types txt,log --exclude-files "*.tmp"
  process --dry-run --verbose --batch-size 5
		`),

		// Default layer parameters (command-specific)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"input-dir",
				parameters.ParameterTypeString,
				parameters.WithHelp("Input directory to process"),
				parameters.WithDefault("./input"),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose logging"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"dry-run",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Show what would be processed without doing it"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"recursive",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Process directories recursively"),
				parameters.WithDefault(true),
			),
		),

		// Add all our custom layers
		cmds.WithLayersList(
			dbLayer,
			processingLayer,
			outputLayer,
			glazedLayer,
		),
	)

	return &FileProcessorCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "file-processor",
		Short: "File processing tool with layered configuration",
		Long: `
A demonstration tool showing how to organize parameters into layers
and load configuration from multiple sources using middleware.
		`,
	}

	processorCmd, err := NewFileProcessorCommand()
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
