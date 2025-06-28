# Tutorial 1: Getting Started with Glazed Commands

This directory contains the complete working code for Tutorial 1, including exercises and solutions.

## Structure

- `cmd/bare/` - BareCommand example
- `cmd/writer/` - WriterCommand example  
- `cmd/glaze/` - GlazeCommand example
- `exercises/` - Practice exercises
- `solutions/` - Exercise solutions

## Running the Examples

```bash
# Install dependencies
go mod init tutorial-01
go get github.com/go-go-golems/glazed

# Test BareCommand
go run cmd/bare/main.go greet --name Alice

# Test WriterCommand
go run cmd/writer/main.go greet --name Bob --prefix ">>> "

# Test GlazeCommand with different outputs
go run cmd/glaze/main.go greet --name Charlie --language spanish
go run cmd/glaze/main.go greet --name Dave --output json --include-time
```

## Exercise

Try implementing the file-info command as described in the tutorial. Solutions are provided in the `solutions/` directory, but try it yourself first!
