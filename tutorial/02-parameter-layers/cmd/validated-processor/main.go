package main

import (
	"context"
	"fmt"
	"os"

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
