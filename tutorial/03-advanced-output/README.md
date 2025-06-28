# Tutorial 3: Advanced Output and Data Processing

This tutorial demonstrates advanced features of the glazed CLI framework including complex data processing, custom templates, and sophisticated output formatting.

## Structure

```
03-advanced-output/
├── README.md                      # This file
├── go.mod                        # Module definition with glazed dependency
├── cmd/                          # Example commands
│   ├── analyzer/                 # Complex data analyzer
│   │   └── main.go
│   └── templated-analyzer/       # Template-based reports
│       └── main.go
├── exercises/                    # Tutorial exercises
│   └── log-analyzer/            # Exercise directory
└── solutions/                   # Solutions to exercises
    └── log-analyzer/            # Complete log analyzer implementation
        ├── main.go
        └── go.mod
```

## Examples

### 1. Data Analyzer (`cmd/analyzer`)

Demonstrates complex data processing with nested structures:

```bash
# Basic summary analysis
go run cmd/analyzer/main.go analyze --analysis summary

# Multiple analysis types with JSON output
go run cmd/analyzer/main.go analyze --analysis summary,customer,product --output json

# Geographic analysis with grouping
go run cmd/analyzer/main.go analyze --analysis geography --group-by state --output table

# Filtered customer analysis
go run cmd/analyzer/main.go analyze --analysis customer --filter-field segment --filter-value premium --limit 3

# Time-based analysis
go run cmd/analyzer/main.go analyze --analysis time --group-by month
```

### 2. Templated Analyzer (`cmd/templated-analyzer`)

Shows custom template usage for formatted reports:

```bash
# Executive summary report
go run cmd/templated-analyzer/main.go report --template executive

# Dashboard view with metadata
go run cmd/templated-analyzer/main.go report --template dashboard --show-metadata

# Detailed breakdown as CSV
go run cmd/templated-analyzer/main.go report --template detailed --output csv

# Custom template example
go run cmd/templated-analyzer/main.go report --template custom --custom-template "{{.Title}}: ${{.Summary.total_revenue}}"
```

### 3. Log Analyzer (`solutions/log-analyzer`)

Complete web server log analysis tool:

```bash
cd solutions/log-analyzer

# Traffic analysis by hour
go run main.go analyze-logs --analysis traffic --group-by hour

# Security analysis
go run main.go analyze-logs --analysis security --template dashboard

# Error analysis with filtering
go run main.go analyze-logs --analysis errors --filter-status 400-499 --limit 10

# Performance analysis
go run main.go analyze-logs --analysis performance,summary --output json
```

## Key Concepts Demonstrated

### Complex Data Structures
- Nested JSON processing with embedded objects and arrays
- Data aggregation across multiple dimensions
- Custom filtering and sorting logic
- Handling missing or malformed data

### Template Systems
- Go template integration with glazed output
- Custom template functions and helpers
- Conditional rendering based on data
- Multiple output formats from single data source

### Advanced Output Features
- Custom formatters for different audiences
- Structured data with flexible presentation
- Metadata inclusion and contextual information
- Dashboard-style visual formatting

### Data Processing Patterns
- Streaming data processing
- Memory-efficient handling of large datasets
- Business logic separation from output formatting
- Error handling and data validation

## Prerequisites

- Completion of Tutorials 1 and 2
- Basic understanding of Go templates
- Familiarity with JSON and data processing concepts

## Building and Testing

```bash
# Install dependencies
go mod tidy

# Test the analyzer
go run cmd/analyzer/main.go analyze --help

# Test the templated analyzer  
go run cmd/templated-analyzer/main.go report --help

# Test the log analyzer solution
cd solutions/log-analyzer
go mod tidy
go run main.go analyze-logs --help
```

## Learning Objectives

By completing this tutorial, you will understand:

1. **Complex Data Processing**: How to work with nested data structures and perform sophisticated analysis
2. **Template Integration**: How to use Go templates with glazed for custom output formatting
3. **Output Customization**: How to create different views of the same data for different audiences
4. **Performance Considerations**: How to handle large datasets efficiently
5. **Real-world Applications**: How to build practical tools like log analyzers

## Next Steps

Continue to Tutorial 4: Production-Ready Tools to learn about:
- Error handling and logging strategies
- Testing frameworks and methodologies
- Performance optimization techniques
- Deployment and distribution considerations
