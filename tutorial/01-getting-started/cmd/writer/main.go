package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

// GreetCommand demonstrates WriterCommand for custom output control
type GreetCommand struct {
	*cmds.CommandDescription
}

type GreetSettings struct {
	Name   string `glazed.parameter:"name"`
	Prefix string `glazed.parameter:"prefix"`
}

var _ cmds.WriterCommand = &GreetCommand{}

func (c *GreetCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	settings := &GreetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Write to the provided writer with custom prefix
	_, err := fmt.Fprintf(w, "%sHello, %s!\n", settings.Prefix, settings.Name)
	return err
}

func NewGreetCommand() (*GreetCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"greet",
		cmds.WithShort("Greet someone with custom prefix"),
		cmds.WithLong("A greeting command that demonstrates WriterCommand with custom output formatting."),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the person to greet"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"prefix",
				parameters.ParameterTypeString,
				parameters.WithHelp("Prefix to add before the greeting"),
				parameters.WithDefault(""),
			),
		),
	)

	return &GreetCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "writer-example",
		Short: "WriterCommand example",
	}

	greetCmd, err := NewGreetCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Build cobra command from WriterCommand using glazed CLI
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
