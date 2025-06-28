---
Title: Parameter Management and Layers
Slug: parameter-management-and-layers
Short: Master parameter organization, validation, and configuration management with layers
Topics:
- parameters
- layers
- middleware
- configuration
- tutorial
Commands:
Flags:
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Parameter Management and Layers

This tutorial teaches you how to organize parameters effectively using glazed's layer system and how to use middleware for flexible configuration management from multiple sources.

## Learning Objectives

- Understand parameter layers and their organization
- Create custom parameter layers for logical grouping
- Use middleware to load configuration from multiple sources
- Implement parameter validation and type conversion
- Handle complex parameter scenarios

## Prerequisites

- Completed Tutorial 1: Getting Started with Glazed Commands
- Understanding of Go structs and interfaces
- Basic knowledge of configuration management concepts

## Tutorial Overview

We'll build a "file-processor" tool that demonstrates:
- Multiple parameter layers (database, processing, output)
- Configuration loading from files, environment variables, and flags
- Complex parameter validation
- Middleware chains for flexible configuration

## Setting Up

```bash
mkdir glazed-tutorial-02
cd glazed-tutorial-02
go mod init glazed-tutorial-02
go get github.com/go-go-golems/glazed
```

## Part 1: Understanding Parameter Layers

Parameter layers help organize related parameters into logical groups. This makes your CLI more maintainable and user-friendly.

### Creating Custom Layers

Create `cmd/processor/main.go`:

```go
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
			types.MRP("worker_id", (i % processingSettings.Workers) + 1),
			types.MRP("batch", i/processingSettings.BatchSize + 1),
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
			parameters.ParameterTypeSecret,
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
```

## Part 2: Configuration Loading with Middleware

Now let's add middleware support to load configuration from multiple sources.

Add to the end of `cmd/processor/main.go`:

```go
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
```

### Test Layer Organization

```bash
# See all available flags organized by layer
go run cmd/processor/main.go process --help

# Test with basic parameters
mkdir -p input output
echo "test content" > input/test.txt
go run cmd/processor/main.go process --input-dir ./input --verbose

# Test with multiple layers
go run cmd/processor/main.go process \
  --input-dir ./input \
  --workers 2 \
  --batch-size 5 \
  --host database.example.com \
  --destination ./output \
  --format csv \
  --output table
```

## Part 3: Advanced Configuration with Files and Environment

Create configuration files to demonstrate loading from multiple sources.

Create `config/database.yaml`:

```yaml
host: prod-database.example.com
port: 5433
database: production_fileprocessor
username: prod_user
password: super_secret_password
```

Create `config/processing.yaml`:

```yaml
batch-size: 50
workers: 8
file-types: ["txt", "json", "csv", "log"]
exclude-files: ["*.tmp", "*.bak", ".DS_Store"]
max-file-size: 104857600  # 100MB
```

Create `config/app.yaml`:

```yaml
input-dir: /data/input
verbose: true
dry-run: false
recursive: true
```

### Enhanced Command with Configuration Loading

Create `cmd/advanced-processor/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/spf13/cobra"
)

type AdvancedProcessorCommand struct {
	*cmds.CommandDescription
}

// Use the same settings structs from the previous example
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

type DefaultSettings struct {
	InputDir  string `glazed.parameter:"input-dir"`
	Verbose   bool   `glazed.parameter:"verbose"`
	DryRun    bool   `glazed.parameter:"dry-run"`
	Recursive bool   `glazed.parameter:"recursive"`
}

var _ cmds.BareCommand = &AdvancedProcessorCommand{}

func (c *AdvancedProcessorCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
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

	// Display configuration from all sources
	fmt.Println("=== Configuration Summary ===")
	fmt.Printf("Input Directory: %s\n", defaultSettings.InputDir)
	fmt.Printf("Verbose: %t\n", defaultSettings.Verbose)
	fmt.Printf("Dry Run: %t\n", defaultSettings.DryRun)
	fmt.Printf("Recursive: %t\n", defaultSettings.Recursive)
	fmt.Println()

	fmt.Println("=== Database Configuration ===")
	fmt.Printf("Host: %s\n", dbSettings.Host)
	fmt.Printf("Port: %d\n", dbSettings.Port)
	fmt.Printf("Database: %s\n", dbSettings.Database)
	fmt.Printf("Username: %s\n", dbSettings.Username)
	fmt.Printf("Password: %s\n", maskPassword(dbSettings.Password))
	fmt.Println()

	fmt.Println("=== Processing Configuration ===")
	fmt.Printf("Batch Size: %d\n", processingSettings.BatchSize)
	fmt.Printf("Workers: %d\n", processingSettings.Workers)
	fmt.Printf("File Types: %v\n", processingSettings.FileTypes)
	fmt.Printf("Exclude Files: %v\n", processingSettings.ExcludeFiles)
	fmt.Printf("Max File Size: %d\n", processingSettings.MaxFileSize)

	return nil
}

func maskPassword(password string) string {
	if password == "" {
		return "<not set>"
	}
	if len(password) <= 2 {
		return "***"
	}
	return password[:2] + "***"
}

// NewAdvancedProcessorCommand creates command with middleware support
func NewAdvancedProcessorCommand() (*AdvancedProcessorCommand, error) {
	// Create the same layers as before
	dbLayer, err := layers.NewParameterLayer("database", "Database Configuration")
	if err != nil {
		return nil, err
	}

	dbLayer.AddFlags(
		parameters.NewParameterDefinition("host", parameters.ParameterTypeString, parameters.WithDefault("localhost")),
		parameters.NewParameterDefinition("port", parameters.ParameterTypeInteger, parameters.WithDefault(5432)),
		parameters.NewParameterDefinition("database", parameters.ParameterTypeString, parameters.WithDefault("fileprocessor")),
		parameters.NewParameterDefinition("username", parameters.ParameterTypeString, parameters.WithDefault("admin")),
		parameters.NewParameterDefinition("password", parameters.ParameterTypeSecret, parameters.WithDefault("")),
	)

	processingLayer, err := layers.NewParameterLayer("processing", "Processing Configuration")
	if err != nil {
		return nil, err
	}

	processingLayer.AddFlags(
		parameters.NewParameterDefinition("batch-size", parameters.ParameterTypeInteger, parameters.WithDefault(10)),
		parameters.NewParameterDefinition("workers", parameters.ParameterTypeInteger, parameters.WithDefault(4)),
		parameters.NewParameterDefinition("file-types", parameters.ParameterTypeStringList, parameters.WithDefault([]string{})),
		parameters.NewParameterDefinition("exclude-files", parameters.ParameterTypeStringList, parameters.WithDefault([]string{})),
		parameters.NewParameterDefinition("max-file-size", parameters.ParameterTypeInteger, parameters.WithDefault(0)),
	)

	cmdDesc := cmds.NewCommandDescription(
		"process",
		cmds.WithShort("Advanced file processor with configuration loading"),
		cmds.WithLong(`
Advanced file processor demonstrating configuration loading from multiple sources.

Configuration is loaded in this order (later sources override earlier ones):
1. Parameter defaults
2. Configuration files (if specified)
3. Environment variables (PROCESSOR_*)
4. Command line flags

Examples:
  # Use default configuration
  process

  # Load configuration from files
  process --load-parameters-from-file config/app.yaml

  # Override with environment variables
  PROCESSOR_WORKERS=16 PROCESSOR_VERBOSE=true process

  # Mix of file and command line
  process --load-parameters-from-file config/app.yaml --workers 8
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition("input-dir", parameters.ParameterTypeString, parameters.WithDefault("./input")),
			parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool, parameters.WithDefault(false)),
			parameters.NewParameterDefinition("dry-run", parameters.ParameterTypeBool, parameters.WithDefault(false)),
			parameters.NewParameterDefinition("recursive", parameters.ParameterTypeBool, parameters.WithDefault(true)),
		),

		cmds.WithLayersList(dbLayer, processingLayer),
	)

	return &AdvancedProcessorCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "advanced-processor",
		Short: "Advanced file processor with configuration management",
	}

	processorCmd, err := NewAdvancedProcessorCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// For this example, we'll use the runner package to demonstrate
	// programmatic execution with middleware
	
	// You can also use CLI integration:
	cobraCmd, err := cli.BuildCobraCommandFromCommand(processorCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	// Add a command that demonstrates programmatic execution with middleware
	demoCmd := &cobra.Command{
		Use:   "demo-middleware",
		Short: "Demonstrate middleware configuration loading",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			
			// Run with middleware chain
			err := runner.ParseAndRun(ctx, processorCmd,
				[]runner.ParseOption{
					// Load from environment with prefix
					runner.WithEnvMiddleware("PROCESSOR_"),
					// Could add file loading here too
					// runner.WithLoadParametersFromFile("config/app.yaml"),
				},
				[]runner.RunOption{},
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(demoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

### Test Configuration Loading

```bash
# Test environment variable loading
PROCESSOR_WORKERS=16 PROCESSOR_VERBOSE=true go run cmd/advanced-processor/main.go demo-middleware

# Test command line override
PROCESSOR_WORKERS=16 go run cmd/advanced-processor/main.go process --workers 8 --verbose

# Create and test config file
mkdir -p config
echo "verbose: true
workers: 12
input-dir: /tmp/test" > config/test.yaml

go run cmd/advanced-processor/main.go process --load-parameters-from-file config/test.yaml
```

## Part 4: Parameter Validation and Custom Types

Let's add some parameter validation to ensure our configuration makes sense.

Create `cmd/validated-processor/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ValidatedProcessorCommand struct {
	*cmds.CommandDescription
}

type ValidatedSettings struct {
	InputDir    string `glazed.parameter:"input-dir"`
	Workers     int    `glazed.parameter:"workers"`
	BatchSize   int    `glazed.parameter:"batch-size"`
	MaxFileSize int    `glazed.parameter:"max-file-size"`
	Verbose     bool   `glazed.parameter:"verbose"`
}

var _ cmds.BareCommand = &ValidatedProcessorCommand{}

func (c *ValidatedProcessorCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ValidatedSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Validate configuration
	if err := c.validateSettings(settings); err != nil {
		return err
	}

	fmt.Println("✅ Configuration validated successfully!")
	fmt.Printf("Input Directory: %s\n", settings.InputDir)
	fmt.Printf("Workers: %d\n", settings.Workers)
	fmt.Printf("Batch Size: %d\n", settings.BatchSize)
	fmt.Printf("Max File Size: %d bytes\n", settings.MaxFileSize)

	return nil
}

func (c *ValidatedProcessorCommand) validateSettings(settings *ValidatedSettings) error {
	// Validate input directory exists
	if _, err := os.Stat(settings.InputDir); os.IsNotExist(err) {
		return errors.Errorf("input directory does not exist: %s", settings.InputDir)
	}

	// Validate workers count
	if settings.Workers < 1 || settings.Workers > 100 {
		return errors.Errorf("workers must be between 1 and 100, got: %d", settings.Workers)
	}

	// Validate batch size
	if settings.BatchSize < 1 || settings.BatchSize > 1000 {
		return errors.Errorf("batch-size must be between 1 and 1000, got: %d", settings.BatchSize)
	}

	// Validate max file size
	if settings.MaxFileSize < 0 {
		return errors.Errorf("max-file-size cannot be negative, got: %d", settings.MaxFileSize)
	}

	// Logical validation: batch size should be reasonable relative to workers
	if settings.BatchSize < settings.Workers && settings.Workers > 1 {
		if settings.Verbose {
			fmt.Printf("⚠️  Warning: batch size (%d) is smaller than worker count (%d)\n", 
				settings.BatchSize, settings.Workers)
		}
	}

	return nil
}

func NewValidatedProcessorCommand() (*ValidatedProcessorCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"validate",
		cmds.WithShort("File processor with parameter validation"),
		cmds.WithLong(`
A file processor that demonstrates parameter validation.

This command validates:
- Input directory exists
- Worker count is reasonable (1-100)
- Batch size is reasonable (1-1000)
- File size limits are not negative
- Logical relationships between parameters

Examples:
  validate --input-dir ./data --workers 4 --batch-size 10
  validate --input-dir /nonexistent  # Will fail validation
  validate --workers 0               # Will fail validation
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"input-dir",
				parameters.ParameterTypeString,
				parameters.WithHelp("Input directory to process (must exist)"),
				parameters.WithDefault("./input"),
			),
			parameters.NewParameterDefinition(
				"workers",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of worker processes (1-100)"),
				parameters.WithDefault(4),
			),
			parameters.NewParameterDefinition(
				"batch-size",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of files per batch (1-1000)"),
				parameters.WithDefault(10),
			),
			parameters.NewParameterDefinition(
				"max-file-size",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum file size in bytes (0 = no limit)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose output including warnings"),
				parameters.WithDefault(false),
			),
		),
	)

	return &ValidatedProcessorCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "validated-processor",
		Short: "File processor with validation",
	}

	validatedCmd, err := NewValidatedProcessorCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(validatedCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

### Test Parameter Validation

```bash
# Create test directory and test successful validation
mkdir -p input
go run cmd/validated-processor/main.go validate --input-dir ./input --workers 4 --batch-size 10

# Test validation failures
go run cmd/validated-processor/main.go validate --input-dir /nonexistent
go run cmd/validated-processor/main.go validate --workers 0
go run cmd/validated-processor/main.go validate --batch-size 1001

# Test warnings
go run cmd/validated-processor/main.go validate --workers 8 --batch-size 2 --verbose
```

## Key Concepts Learned

### Parameter Layers
- **Organization**: Group related parameters into logical layers
- **Separation of Concerns**: Database, processing, output, etc.
- **Maintainability**: Easier to manage complex applications

### Configuration Management
- **Multiple Sources**: Files, environment variables, command line
- **Precedence**: Later sources override earlier ones
- **Middleware**: Flexible configuration loading chains

### Parameter Validation
- **Type Safety**: Automatic type conversion and validation
- **Business Logic**: Custom validation for your domain
- **User Experience**: Clear error messages and warnings

### Best Practices
- Use descriptive layer names and help text
- Provide sensible defaults
- Validate parameters early
- Group related parameters together
- Use environment variables for deployment-specific config

## What's Next

In the next tutorial, you'll learn about:
- Advanced data processing techniques
- Custom output formatters
- Template systems for complex output
- Integration with external data sources

## Exercise

Create a "backup-tool" command with these requirements:

1. **Layers**:
   - Source: source directories, exclusion patterns
   - Destination: backup location, compression options
   - Schedule: backup frequency, retention policy

2. **Configuration Loading**:
   - Load from `backup-config.yaml`
   - Override with `BACKUP_*` environment variables
   - Command line flags take precedence

3. **Validation**:
   - Source directories must exist
   - Destination must be writable
   - Retention days must be positive
   - Compression level must be valid (0-9)

4. **Output**:
   - Use GlazeCommand to show backup plan
   - Include source paths, sizes, and estimated backup time

Test your implementation with various configuration sources and validation scenarios!
