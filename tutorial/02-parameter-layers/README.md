# Tutorial 2: Parameter Management and Layers

This tutorial demonstrates advanced parameter organization using glazed's layer system and configuration management with middleware.

## Overview

Learn how to:
- Create custom parameter layers for logical grouping
- Load configuration from multiple sources (files, environment, CLI)
- Implement parameter validation and type conversion
- Build production-ready tools with complex configuration

## Examples

### 1. Basic Layered Parameters (`cmd/processor/`)
Demonstrates organizing parameters into logical layers:
- Database configuration
- Processing settings
- Output configuration
- Glazed output formatting

### 2. Advanced Configuration (`cmd/advanced-processor/`)
Shows configuration loading from multiple sources with proper precedence.

### 3. Parameter Validation (`cmd/validated-processor/`)
Implements comprehensive parameter validation with clear error messages.

### 4. Production Tool (`solutions/backup-tool/`)
A complete backup tool showing practical application of all concepts.

## Quick Start

```bash
# Install dependencies
go mod tidy

# Test basic layered parameters
mkdir -p input output
echo "test content" > input/test.txt
go run cmd/processor/main.go process --input-dir ./input --verbose

# Test with different layers
go run cmd/processor/main.go process \
  --input-dir ./input \
  --workers 2 \
  --batch-size 5 \
  --host database.example.com \
  --destination ./output \
  --format csv \
  --output table

# Test configuration loading
go run cmd/advanced-processor/main.go process --load-parameters-from-file config/app.yaml

# Test environment variables
PROCESSOR_WORKERS=16 PROCESSOR_VERBOSE=true go run cmd/advanced-processor/main.go process

# Test validation
mkdir -p input
go run cmd/validated-processor/main.go validate --input-dir ./input --workers 4
go run cmd/validated-processor/main.go validate --workers 0  # Will fail

# Test backup tool
cd solutions/backup-tool
mkdir -p source backups
echo "important data" > source/data.txt
go run main.go backup --source-dir ./source --backup-dir ./backups --dry-run
```

## Key Concepts

### Parameter Layers
- **Organization**: Group related parameters logically
- **Maintainability**: Easier to manage complex configurations
- **User Experience**: Clear help organization and documentation

### Configuration Sources
1. Parameter defaults
2. Configuration files (YAML/JSON)
3. Environment variables
4. Command line flags

### Validation
- Type safety with automatic conversion
- Business logic validation
- Clear error messages and warnings
- Logical relationship validation

## Configuration Files

Sample configuration files are provided in the `config/` directory:
- `app.yaml`: Application-level settings
- `database.yaml`: Database connection settings
- `processing.yaml`: File processing configuration

## Production Usage

The backup-tool example shows how to structure a production tool with:
- Multiple configuration layers
- Comprehensive validation
- Structured output for monitoring
- Dry-run capability for safety
- Clear error handling and user feedback

## Next Steps

After completing this tutorial, explore:
- Advanced output formatting (Tutorial 3)
- Production deployment patterns (Tutorial 4)
- Custom middleware development
- Configuration templating systems
