# Glazed Development Guidelines

## Build and Test Commands
- Build: `make build` - Runs go generate and builds binaries with version info
- Install: `make install` - Builds and installs the glaze binary
- Lint: `make lint` - Runs golangci-lint with verbosity flag
- Lint exhaustive: `make exhaustive` - Runs golangci-lint with exhaustive linter  
- Test all: `make test` - Runs go test ./...
- Test single: `go test -v [package_path] -run TestName`
- Benchmark: `make bench` - Runs benchmarks with -benchmem flag

## Code Style
- Use gofmt (enforced via golangci-lint)
- Import ordering: stdlib first, then third-party packages
- Naming: CamelCase for exported symbols, snake_case for test files
- Types: Use interfaces and strong typing with custom types in pkg/types
- Error handling: Use pkg/errors for wrapping, always propagate with context
- Function options pattern for configuration
- Tests: Use testify/assert and testify/require packages
- Git hooks: Pre-commit and pre-push hooks run via lefthook (see lefthook.yml)

Always run golangci-lint before committing changes (`make lint`).