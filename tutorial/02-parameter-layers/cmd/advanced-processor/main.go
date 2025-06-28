package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
		parameters.NewParameterDefinition("password", parameters.ParameterTypeString, parameters.WithDefault("")),
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
