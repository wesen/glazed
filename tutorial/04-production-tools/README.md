# Tutorial 4: Building Production CLI Tools

This tutorial demonstrates how to build robust, production-ready CLI applications using glazed. You'll learn about error handling, logging, testing strategies, configuration management, and deployment best practices.

## What You'll Learn

- Implement comprehensive error handling and logging
- Create testable command structures
- Build configuration management systems
- Add health checks and monitoring
- Deploy and distribute CLI applications
- Handle performance and security considerations

## Prerequisites

- Completed Tutorials 1, 2, and 3
- Understanding of Go testing frameworks
- Basic knowledge of CI/CD concepts
- Familiarity with deployment strategies

## Getting Started

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run the production processor:**
   ```bash
   go run cmd/production-processor/main.go process --help
   ```

3. **Create a sample log file:**
   ```bash
   cat > sample.log << 'EOF'
   127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
   192.168.1.1 - - [10/Oct/2000:13:55:37 -0700] "POST /form HTTP/1.0" 404 1234
   10.0.0.1 - - [10/Oct/2000:13:55:38 -0700] "GET /index.html HTTP/1.0" 200 5432
   EOF
   ```

4. **Process the log file:**
   ```bash
   go run cmd/production-processor/main.go process --input-file sample.log --parser apache --verbose
   ```

5. **Run health check:**
   ```bash
   go run cmd/production-processor/main.go process --health-check
   ```

6. **Start with metrics:**
   ```bash
   go run cmd/production-processor/main.go process --input-file sample.log --parser apache --start-metrics
   # In another terminal: curl http://localhost:9090/metrics
   ```

## Testing

Run all tests:
```bash
go test ./...
```

Run benchmarks:
```bash
go test -bench=. ./internal/processor/
```

## Architecture

The tutorial follows production-ready patterns:

- **Error Handling**: Custom error types with codes and context
- **Logging**: Structured logging with zerolog
- **Metrics**: Prometheus metrics for monitoring
- **Configuration**: YAML-based config with environment overrides
- **Health Checks**: HTTP endpoints for monitoring
- **Testing**: Unit tests, integration tests, and benchmarks

## Exercise

See the `exercises/` directory for hands-on challenges, and `solutions/` for reference implementations.

## Next Steps

After completing this tutorial, you'll be ready to build production-grade CLI applications with proper error handling, monitoring, and deployment capabilities.
