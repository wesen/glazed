package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// TemplateDemoCommand demonstrates the use of named templates in a GlazeCommand
type TemplateDemoCommand struct {
	cmd *cmds.CommandDescription
}

func NewTemplateDemoCommand() (*TemplateDemoCommand, error) {
	cmd := &cmds.CommandDescription{
		Name:  "template-demo",
		Short: "Demonstrates named templates in a GlazeCommand",
		Long: "This command demonstrates how to use named templates in a GlazeCommand. " +
			"It provides multiple template formats (default, json, markdown) that can be selected using the --glazed-template-name flag.",
		Layers: []layers.Layer{},
		TemplateConfig: &cmds.TemplateConfig{
			Templates: map[string]string{
				"default":  "# Template Demo Results\n{{range .rows}}- Name: {{.name}}, Value: {{.value}}\n{{end}}",
				"json":     "{\n  \"results\": [\n{{range $i, $row := .rows}}    {{if $i}},{{end}}{\n      \"name\": \"{{.name}}\",\n      \"value\": {{.value}}\n    }\n{{end}}  ]\n}",
				"markdown": "# Template Demo Results\n\n| Name | Value |\n|------|-------|\n{{range .rows}}| {{.name}} | {{.value}} |\n{{end}}",
			},
			DefaultTemplate: "default",
		},
	}

	return &TemplateDemoCommand{
		cmd: cmd,
	}, nil
}

func (c *TemplateDemoCommand) Description() *cmds.CommandDescription {
	return c.cmd
}

func (c *TemplateDemoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor) error {

	// Generate some sample data
	items := []struct {
		Name  string
		Value int
	}{
		{"Item 1", 100},
		{"Item 2", 200},
		{"Item 3", 300},
	}

	// Add rows to the processor
	for _, item := range items {
		row := types.NewRow()
		row.Set("name", item.Name)
		row.Set("value", item.Value)
		gp.AddRow(row)
	}

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "template-demo",
		Short: "Example of named templates in GlazeCommands",
	}

	templateDemo, err := NewTemplateDemoCommand()
	if err != nil {
		fmt.Printf("Error creating command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(templateDemo)
	if err != nil {
		fmt.Printf("Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}