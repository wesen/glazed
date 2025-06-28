# Simple Command Example

This example demonstrates how to create a basic glazed command with structured data output.

## Features Demonstrated

- **GlazeCommand**: Outputs structured data that can be formatted as tables, JSON, CSV, etc.
- **Parameter Layers**: Organizes parameters into logical groups
- **Settings Struct**: Clean parameter access using struct tags
- **Glazed Integration**: Automatic support for all glazed output formats and flags

## Running the Example

```bash
# Build and run
go run main.go list-users

# Try different output formats
go run main.go list-users --output json
go run main.go list-users --output csv
go run main.go list-users --output yaml

# Filter and limit results
go run main.go list-users --filter admin
go run main.go list-users --count 2

# Select specific fields
go run main.go list-users --fields username,role

# Verbose output
go run main.go list-users --verbose

# Get help
go run main.go list-users --help
```

## Key Concepts

### Command Structure
- Commands embed `*cmds.CommandDescription`
- Implement one of: `BareCommand`, `WriterCommand`, or `GlazeCommand`
- Use constructor functions to set up parameters and layers

### Parameter Handling
- Define parameters with types, defaults, and help text
- Use struct tags (`glazed.parameter:"name"`) for clean access
- Parameters are automatically validated and parsed

### Layers
- **Default Layer**: Your command-specific parameters
- **Glazed Layer**: Standard output formatting options
- **Custom Layers**: Group related configuration

### Output
- GlazeCommands output rows that are automatically formatted
- Support for tables, JSON, YAML, CSV, and more
- Built-in field selection, filtering, and transformation
