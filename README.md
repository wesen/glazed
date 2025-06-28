# glazed - Framework for Rich Command-Line Tools with Structured Data Output

![](https://img.shields.io/github/license/go-go-golems/glazed)
![](https://img.shields.io/github/actions/workflow/status/go-go-golems/glazed/push.yml?branch=main)

> Build beautiful command-line tools with structured data output, rich help systems, and flexible parameter parsing.

**glazed** is a comprehensive Go framework for building command-line applications that need to handle structured data with rich output formatting, flexible parameter management, and powerful help systems.

## 🚀 What is glazed?

glazed transforms the way you build CLI tools by providing:

- **Three Command Types**: Choose between `BareCommand`, `WriterCommand`, and `GlazeCommand` based on your output needs
- **Structured Data Pipeline**: Built-in support for outputting data as JSON, YAML, CSV, tables, and more
- **Layered Parameter System**: Organize parameters into logical groups with powerful parsing and validation
- **Rich Help System**: Create comprehensive documentation with topics, examples, applications, and tutorials
- **Middleware Architecture**: Process parameters from multiple sources (flags, env vars, config files)
- **Easy Integration**: Works seamlessly with Cobra and can be used programmatically

## 📦 Quick Start

### Installation

```bash
go get github.com/go-go-golems/glazed
```

### Try the glaze CLI Tool

The `glaze` command demonstrates glazed's capabilities:

```bash
# Run the tool
go run ./cmd/glaze

# Process JSON data with table output
echo '{"id": 1, "name": "John", "email": "john@example.com"}' | go run ./cmd/glaze json

# Output as CSV
echo '{"id": 1, "name": "John"}' | go run ./cmd/glaze json --output csv

# Get help on any topic
go run ./cmd/glaze help help-system
```

## 🏗️ Building Your First Command

Here's a simple example of creating a GlazeCommand that outputs structured data:

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
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
    "github.com/spf13/cobra"
)

// UserListCommand demonstrates a GlazeCommand
type UserListCommand struct {
    *cmds.CommandDescription
}

// Settings struct for clean parameter access
type UserListSettings struct {
    Count int  `glazed.parameter:"count"`
    Verbose bool `glazed.parameter:"verbose"`
}

var _ cmds.GlazeCommand = &UserListCommand{}

func (c *UserListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Parse settings from layers
    s := &UserListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Generate sample data
    users := []struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
        Role string `json:"role"`
    }{
        {1, "Alice", "admin"},
        {2, "Bob", "user"},
        {3, "Carol", "editor"},
    }
    
    // Limit results if specified
    if s.Count > 0 && s.Count < len(users) {
        users = users[:s.Count]
    }
    
    // Output as structured rows
    for _, user := range users {
        row := types.NewRowFromStruct(&user, true)
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}

func NewUserListCommand() (*UserListCommand, error) {
    // Create standard glazed output layer
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    // Define command with parameters
    cmdDesc := cmds.NewCommandDescription(
        "list-users",
        cmds.WithShort("List users with optional filtering"),
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "count",
                parameters.ParameterTypeInteger,
                parameters.WithHelp("Maximum number of users to return"),
                parameters.WithDefault(0),
            ),
            parameters.NewParameterDefinition(
                "verbose",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Enable verbose output"),
                parameters.WithDefault(false),
            ),
        ),
        cmds.WithLayersList(glazedLayer),
    )
    
    return &UserListCommand{CommandDescription: cmdDesc}, nil
}

func main() {
    // Create root command
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "Example glazed application",
    }
    
    // Create and add your command
    userCmd, err := NewUserListCommand()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    cobraCmd, err := cli.BuildCobraCommandFromCommand(userCmd)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    rootCmd.AddCommand(cobraCmd)
    
    // Execute
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

Run your command:

```bash
# Table output (default)
./myapp list-users

# JSON output
./myapp list-users --output json

# Limit results
./myapp list-users --count 2 --output yaml
```

## 🎯 Core Concepts

### Command Types

glazed provides three command interfaces for different use cases:

1. **`BareCommand`**: Handle your own output directly
2. **`WriterCommand`**: Write to a provided `io.Writer`
3. **`GlazeCommand`**: Output structured data with automatic formatting

### Parameter Layers

Organize parameters into logical groups:

- **Default Layer**: Your command-specific parameters
- **Glazed Layer**: Standard output formatting options (`--output`, `--fields`, etc.)
- **Custom Layers**: Group related configuration (database, API, etc.)

```go
// Create a custom layer
dbLayer, err := layers.NewParameterLayer(
    "database",
    "Database Configuration",
)
dbLayer.AddFlags(
    parameters.NewParameterDefinition("host", parameters.ParameterTypeString),
    parameters.NewParameterDefinition("port", parameters.ParameterTypeInteger),
)
```

### Parameter Types

Rich type system with validation:

- **Basic**: `String`, `Integer`, `Bool`, `Float`, `Date`
- **Secret**: `Secret` (masks sensitive values)
- **Lists**: `StringList`, `IntegerList`, `FloatList`
- **Choices**: `Choice`, `ChoiceList` (predefined options)
- **Files**: `File`, `FileList`, `StringFromFile`
- **Key-Value**: `KeyValue` (map-like inputs)

### Middleware System

Process parameters from multiple sources:

```go
// Chain middleware for parameter processing
err := middlewares.ExecuteMiddlewares(layers, parsedLayers,
    middlewares.SetFromDefaults(),
    middlewares.UpdateFromEnv("MYAPP"),
    middlewares.LoadParametersFromFile("config.yaml"),
    middlewares.ParseFromCobraCommand(cmd),
)
```

## 🎨 Output Formats

GlazeCommands automatically support multiple output formats:

```bash
# Table (default)
myapp command

# JSON
myapp command --output json

# YAML  
myapp command --output yaml

# CSV
myapp command --output csv

# Select specific fields
myapp command --fields id,name --output json

# Filter and rename columns
myapp command --filter-columns age --rename name:username
```

## 📚 Help System

Create rich documentation with multiple section types:

### General Topics
Comprehensive articles about concepts and features

### Examples  
Specific command usage examples

### Applications
Real-world use cases combining multiple tools

### Tutorials
Step-by-step guides for complex workflows

```bash
# Access help system
myapp help help-system

# List all help topics
myapp help --list

# Get help on specific topics
myapp help parameter-layers
```

## 🔧 Advanced Features

### YAML Command Definitions

Define commands declaratively:

```yaml
name: process-data
short: Process data with filtering options
flags:
  - name: input-file
    type: file
    help: Input data file
    required: true
  - name: format
    type: choice
    choices: [json, yaml, csv]
    default: json
    help: Output format
arguments:
  - name: output-dir
    type: string
    help: Output directory
    required: true
```

### Programmatic Execution

Run commands without CLI:

```go
// Run programmatically
ctx := context.Background()
err := runner.ParseAndRun(ctx, cmd, 
    []runner.ParseOption{
        runner.WithEnvMiddleware("MYAPP_"),
        runner.WithValuesForLayers(map[string]map[string]interface{}{
            "default": {"verbose": true},
        }),
    },
    []runner.RunOption{
        runner.WithWriter(os.Stdout),
    },
)
```

### JSON Schema Generation

Generate schemas for validation and documentation:

```go
schema, err := cmd.Description().ToJsonSchema()
```

## 🛠️ Real-World Applications

glazed powers several production tools:

- **Data Processing Pipelines**: Transform JSON, YAML, CSV data with rich output
- **API Clients**: Build CLI tools that consume APIs and format responses
- **System Administration**: Create tools that gather system info with structured output
- **Configuration Management**: Build tools that validate and transform config files

## 📋 Migration Guide

If you're using the old glazed interface focused on simple data formatting:

### Old Approach (still supported)
```go
// Direct formatter usage
formatter := formatters.NewTableFormatter()
formatter.AddRow(types.NewRow(types.MRP("name", "value")))
output, _ := formatter.Output()
```

### New Approach (recommended)
```go
// Command-based approach with rich features
type MyCommand struct {
    *cmds.CommandDescription
}

func (c *MyCommand) RunIntoGlazeProcessor(ctx context.Context, 
    parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    return gp.AddRow(ctx, types.NewRow(types.MRP("name", "value")))
}
```

## 🤝 Contributing

Contributions are welcome! Please check our [contribution guidelines](CONTRIBUTING.md).

## 📄 License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.

## 📚 Learning Resources

### Comprehensive Tutorial Series

Learn glazed step-by-step with our comprehensive tutorial series:

1. **[Getting Started with Glazed Commands](pkg/doc/tutorials/01-getting-started-with-commands.md)**
   - Learn the three command types (BareCommand, WriterCommand, GlazeCommand)
   - Create your first commands with parameters
   - Understand structured data output
   - [Practice Exercise: File Info Tool](tutorial/01-getting-started/exercises/file-info/)

2. **[Parameter Management and Layers](pkg/doc/tutorials/02-parameter-management-and-layers.md)**
   - Master parameter organization with layers
   - Configuration loading from multiple sources
   - Middleware for flexible parameter processing
   - [Practice Exercise: Backup Tool](tutorial/02-parameter-layers/exercises/backup-tool/)

3. **[Advanced Output and Data Processing](pkg/doc/tutorials/03-advanced-output-and-data-processing.md)**
   - Handle complex nested data structures
   - Create custom templates and formatters
   - Build sophisticated data processing pipelines
   - [Practice Exercise: Log Analyzer](tutorial/03-advanced-output/exercises/log-analyzer/)

4. **[Building Production CLI Tools](pkg/doc/tutorials/04-building-production-cli-tools.md)**
   - Implement robust error handling and logging
   - Add monitoring and health checks
   - Create comprehensive test strategies
   - [Practice Exercise: File Synchronizer](tutorial/04-production-tools/exercises/file-synchronizer/)

### Quick Start Examples

- **[Simple Command Example](examples/simple-command/)** - Complete working example
- **[Tutorial Code](tutorial/)** - All tutorial examples and exercises
- **[Demo Scripts](demos/)** - VHS demonstration tapes

### Help System

glazed includes a rich help system. Access it with:

```bash
# Get help on any topic
go run ./cmd/glaze help help-system

# List all available help topics
go run ./cmd/glaze help --list

# Get help on specific topics
go run ./cmd/glaze help parameter-layers
```

## 🔗 Links

- [API Documentation](https://pkg.go.dev/github.com/go-go-golems/glazed)
- [Tutorial Series](pkg/doc/tutorials/)
- [Examples](examples/)
- [Issue Tracker](https://github.com/go-go-golems/glazed/issues)
