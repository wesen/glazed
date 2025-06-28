package main

import (
	"context"
	"fmt"
	"os"
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

// GreetCommand demonstrates GlazeCommand with structured output
type GreetCommand struct {
	*cmds.CommandDescription
}

type GreetSettings struct {
	Name        string `glazed.parameter:"name"`
	Language    string `glazed.parameter:"language"`
	IncludeTime bool   `glazed.parameter:"include-time"`
}

var _ cmds.GlazeCommand = &GreetCommand{}

func (c *GreetCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &GreetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Generate greeting based on language
	greetings := map[string]string{
		"english": "Hello",
		"spanish": "Hola",
		"french":  "Bonjour",
		"german":  "Hallo",
		"italian": "Ciao",
	}

	greeting, exists := greetings[settings.Language]
	if !exists {
		greeting = "Hello" // default to English
	}

	// Create structured output row
	row := types.NewRow(
		types.MRP("greeting", greeting),
		types.MRP("name", settings.Name),
		types.MRP("language", settings.Language),
		types.MRP("message", fmt.Sprintf("%s, %s!", greeting, settings.Name)),
	)

	// Add timestamp if requested
	if settings.IncludeTime {
		row.Set("timestamp", time.Now().Format(time.RFC3339))
		row.Set("unix_timestamp", time.Now().Unix())
	}

	return gp.AddRow(ctx, row)
}

func NewGreetCommand() (*GreetCommand, error) {
	// Create glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"greet",
		cmds.WithShort("Greet someone in different languages with structured output"),
		cmds.WithLong(`
Greet someone in different languages and output structured data.
Supports multiple output formats through Glazed (JSON, YAML, CSV, table, etc.).

Examples:
  greet --name Alice --language spanish
  greet --name Bob --language french --output json
  greet --name Charlie --include-time --output yaml
  greet --name Dave --fields message,timestamp --output csv
		`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the person to greet"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"language",
				parameters.ParameterTypeString,
				parameters.WithHelp("Language for greeting (english, spanish, french, german, italian)"),
				parameters.WithDefault("english"),
			),
			parameters.NewParameterDefinition(
				"include-time",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include timestamp in output"),
				parameters.WithDefault(false),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &GreetCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "glaze-example",
		Short: "GlazeCommand example with structured output",
	}

	greetCmd, err := NewGreetCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(greetCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
