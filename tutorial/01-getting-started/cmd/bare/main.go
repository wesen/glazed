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

// GreetCommand demonstrates the most basic BareCommand
type GreetCommand struct {
	*cmds.CommandDescription
}

// Settings for clean parameter access
type GreetSettings struct {
	Name string `glazed.parameter:"name"`
}

var _ cmds.BareCommand = &GreetCommand{}

func (c *GreetCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	// Parse settings from layers
	settings := &GreetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Simple output - just print to stdout
	fmt.Printf("Hello, %s!\n", settings.Name)
	return nil
}

func NewGreetCommand() (*GreetCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"greet",
		cmds.WithShort("Greet someone by name"),
		cmds.WithLong("A simple greeting command that demonstrates BareCommand usage."),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the person to greet"),
				parameters.WithRequired(true),
			),
		),
	)

	return &GreetCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "bare-example",
		Short: "BareCommand example",
	}

	greetCmd, err := NewGreetCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Build cobra command from BareCommand using glazed CLI
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
