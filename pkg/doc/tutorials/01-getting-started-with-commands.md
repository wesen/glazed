---
Title: Getting Started with Glazed Commands
Slug: getting-started-with-commands
Short: Learn to create your first glazed commands with different output types
Topics:
- commands
- tutorial
- beginner
Commands:
Flags:
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Getting Started with Glazed Commands

Welcome to glazed! This tutorial will guide you through creating your first commands using the glazed framework. By the end of this tutorial, you'll understand the three types of commands and how to choose the right one for your needs.

## Learning Objectives

- Understand the three command types: BareCommand, WriterCommand, and GlazeCommand
- Create a simple command with parameters
- Learn basic parameter types and validation
- Output data in different formats

## Prerequisites

- Basic Go programming knowledge
- Go 1.19 or later installed
- Familiarity with command-line interfaces

## Tutorial Overview

We'll build a simple "greet" command that demonstrates all three command types. This command will:
- Accept a name parameter
- Support different greeting styles
- Show how each command type handles output differently

## Setting Up

First, let's create our tutorial project:

```bash
mkdir glazed-tutorial-01
cd glazed-tutorial-01
go mod init glazed-tutorial-01
go get github.com/go-go-golems/glazed
```

## Part 1: BareCommand - Direct Output Control

A `BareCommand` gives you complete control over output. It's perfect when you need to handle output yourself or integrate with existing systems.

### Creating Your First BareCommand

Create `cmd/bare/main.go`:

```go
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

// GreetBareCommand demonstrates a BareCommand
type GreetBareCommand struct {
	*cmds.CommandDescription
}

// Settings for our command
type GreetSettings struct {
	Name      string `glazed.parameter:"name"`
	Uppercase bool   `glazed.parameter:"uppercase"`
	Repeat    int    `glazed.parameter:"repeat"`
}

// Ensure interface implementation
var _ cmds.BareCommand = &GreetBareCommand{}

// Run implements the BareCommand interface
func (c *GreetBareCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	// Parse settings from layers
	s := &GreetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create greeting message
	greeting := fmt.Sprintf("Hello, %s!", s.Name)
	
	if s.Uppercase {
		greeting = strings.ToUpper(greeting)
	}

	// Output the greeting (repeated if requested)
	for i := 0; i < s.Repeat; i++ {
		fmt.Println(greeting)
	}

	return nil
}

func NewGreetBareCommand() (*GreetBareCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"greet",
		cmds.WithShort("Greet someone"),
		cmds.WithLong("A simple greeting command that demonstrates BareCommand usage"),
		
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name to greet"),
				parameters.WithDefault("World"),
			),
			parameters.NewParameterDefinition(
				"uppercase",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Convert greeting to uppercase"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"repeat",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of times to repeat the greeting"),
				parameters.WithDefault(1),
			),
		),
	)

	return &GreetBareCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "bare-greet",
		Short: "Bare command greeting example",
	}

	greetCmd, err := NewGreetBareCommand()
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
```

Don't forget to add the import for strings:

```go
import (
	"context"
	"fmt"
	"os"
	"strings"  // Add this import
	// ... other imports
)
```

### Test Your BareCommand

```bash
go run cmd/bare/main.go greet --name Alice
go run cmd/bare/main.go greet --name Bob --uppercase --repeat 3
```

## Part 2: WriterCommand - Flexible Output Destination

A `WriterCommand` writes to an `io.Writer`, making it easy to redirect output to files, networks, or other destinations.

Create `cmd/writer/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

type GreetWriterCommand struct {
	*cmds.CommandDescription
}

type GreetSettings struct {
	Name      string `glazed.parameter:"name"`
	Uppercase bool   `glazed.parameter:"uppercase"`
	Repeat    int    `glazed.parameter:"repeat"`
	Prefix    string `glazed.parameter:"prefix"`
}

var _ cmds.WriterCommand = &GreetWriterCommand{}

func (c *GreetWriterCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &GreetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	greeting := fmt.Sprintf("%sHello, %s!", s.Prefix, s.Name)
	
	if s.Uppercase {
		greeting = strings.ToUpper(greeting)
	}

	for i := 0; i < s.Repeat; i++ {
		_, err := fmt.Fprintf(w, "%s\n", greeting)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewGreetWriterCommand() (*GreetWriterCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"greet",
		cmds.WithShort("Greet someone with writer output"),
		cmds.WithLong("A greeting command that demonstrates WriterCommand usage"),
		
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name to greet"),
				parameters.WithDefault("World"),
			),
			parameters.NewParameterDefinition(
				"uppercase",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Convert greeting to uppercase"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"repeat",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of times to repeat the greeting"),
				parameters.WithDefault(1),
			),
			parameters.NewParameterDefinition(
				"prefix",
				parameters.ParameterTypeString,
				parameters.WithHelp("Prefix for the greeting"),
				parameters.WithDefault(""),
			),
		),
	)

	return &GreetWriterCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "writer-greet",
		Short: "Writer command greeting example",
	}

	greetCmd, err := NewGreetWriterCommand()
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
```

### Test Your WriterCommand

```bash
go run cmd/writer/main.go greet --name Charlie --prefix ">>> "
```

## Part 3: GlazeCommand - Structured Data Output

A `GlazeCommand` is the most powerful option, automatically supporting multiple output formats like JSON, YAML, CSV, and tables.

Create `cmd/glaze/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
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

type GreetGlazeCommand struct {
	*cmds.CommandDescription
}

type GreetSettings struct {
	Name         string `glazed.parameter:"name"`
	Language     string `glazed.parameter:"language"`
	IncludeTime  bool   `glazed.parameter:"include-time"`
	PersonalInfo bool   `glazed.parameter:"personal-info"`
}

var _ cmds.GlazeCommand = &GreetGlazeCommand{}

func (c *GreetGlazeCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GreetSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Different greetings by language
	greetings := map[string]string{
		"english": "Hello",
		"spanish": "Hola",
		"french":  "Bonjour",
		"german":  "Hallo",
		"italian": "Ciao",
	}

	greeting, exists := greetings[strings.ToLower(s.Language)]
	if !exists {
		greeting = greetings["english"]
	}

	// Create structured data row
	rowData := map[string]interface{}{
		"greeting": greeting,
		"name":     s.Name,
		"language": s.Language,
		"message":  fmt.Sprintf("%s, %s!", greeting, s.Name),
	}

	if s.IncludeTime {
		rowData["timestamp"] = time.Now().Format(time.RFC3339)
	}

	if s.PersonalInfo {
		rowData["formal_greeting"] = fmt.Sprintf("Good day, %s. It is a pleasure to meet you.", s.Name)
		rowData["casual_greeting"] = fmt.Sprintf("Hey %s, what's up?", s.Name)
	}

	row := types.NewRowFromMap(rowData)
	return gp.AddRow(ctx, row)
}

func NewGreetGlazeCommand() (*GreetGlazeCommand, error) {
	// Add glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"greet",
		cmds.WithShort("Greet someone with structured output"),
		cmds.WithLong(`
A greeting command that demonstrates GlazeCommand usage.
Supports multiple output formats and languages.

Examples:
  greet --name Alice --language spanish --output json
  greet --name Bob --include-time --output yaml
  greet --name Carol --personal-info --fields message,formal_greeting
		`),
		
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name to greet"),
				parameters.WithDefault("World"),
			),
			parameters.NewParameterDefinition(
				"language",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Language for greeting"),
				parameters.WithChoices("english", "spanish", "french", "german", "italian"),
				parameters.WithDefault("english"),
			),
			parameters.NewParameterDefinition(
				"include-time",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include timestamp in output"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"personal-info",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include additional personal greeting variations"),
				parameters.WithDefault(false),
			),
		),
		
		// Add glazed layers for rich output options
		cmds.WithLayersList(glazedLayer),
	)

	return &GreetGlazeCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "glaze-greet",
		Short: "Glaze command greeting example",
	}

	greetCmd, err := NewGreetGlazeCommand()
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
```

### Test Your GlazeCommand

```bash
# Table output (default)
go run cmd/glaze/main.go greet --name Alice --language spanish

# JSON output
go run cmd/glaze/main.go greet --name Bob --include-time --output json

# Select specific fields
go run cmd/glaze/main.go greet --name Carol --personal-info --fields message,formal_greeting

# YAML output
go run cmd/glaze/main.go greet --name Dave --language french --output yaml
```

## Key Concepts Learned

### Command Types

1. **BareCommand**: Full control over output, simple implementation
2. **WriterCommand**: Flexible output destination, testable with custom writers
3. **GlazeCommand**: Structured data with automatic formatting support

### Parameter Management

- Use `glazed.parameter` struct tags for clean parameter access
- Parameter types include String, Integer, Bool, Choice, and more
- Default values and help text improve user experience

### When to Use Each Type

- **BareCommand**: Simple utilities, integration with existing systems
- **WriterCommand**: When you need to control output destination
- **GlazeCommand**: Data-rich applications, when you want multiple output formats

## What's Next

In the next tutorial, you'll learn about:
- Parameter layers and organization
- Middleware for configuration management
- Custom parameter types
- Advanced validation

## Exercise

Create a "file-info" command using any command type that:
1. Accepts a file path as an argument
2. Has a flag for including hidden information
3. Outputs file size, modification time, and permissions
4. Test it with different files on your system

Try implementing it with different command types to see the differences!
