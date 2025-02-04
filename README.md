# TerraOrbis Common Library

A foundational Go library providing essential components for building robust microservices in the TerraOrbis platform.

## Overview

This library provides a set of common utilities and patterns used across TerraOrbis services, ensuring consistency, reliability, and maintainability. This is a private library owned by StackCatalyst and is not intended for public distribution.

## Components

### Error Handling (`pkg/errors`)

Standardized error handling with support for:
- Error codes and types
- Error wrapping and context
- Error chain inspection
- Structured error messages

```go
// Example usage
err := errors.New(errors.ErrNotFound, "user not found")
wrappedErr := errors.Wrap(err, errors.ErrInternal, "failed to process request")
```

### Structured Logging (`pkg/logging`)

Built on top of Uber's zap logger, providing:
- Structured JSON logging
- Log levels (Debug, Info, Warn, Error)
- Context-aware logging
- Performance-oriented design

```go
// Example usage
logger, _ := logging.New(logging.DefaultConfig())
logger.Info("server started", zap.Int("port", 8080))
logger.Error("request failed", zap.Error(err))
```

### Configuration Management (`pkg/config`)

Flexible configuration system using Viper with support for:
- Multiple configuration sources (files, environment variables)
- Multiple formats (YAML, JSON)
- Dynamic configuration reloading
- Type-safe configuration access
- Environment variable overrides

```go
// Example usage
cfg, _ := config.New(config.DefaultOptions())
dbHost := cfg.GetString("database.host")
serverPort := cfg.GetInt("server.port")
```

## Internal Usage

Import the required packages:

```go
import (
    "github.com/StackCatalyst/common-lib/pkg/errors"
    "github.com/StackCatalyst/common-lib/pkg/logging"
    "github.com/StackCatalyst/common-lib/pkg/config"
)
```

### Example

```go
package main

import (
    "github.com/StackCatalyst/common-lib/pkg/config"
    "github.com/StackCatalyst/common-lib/pkg/logging"
)

func main() {
    // Initialize configuration
    cfg, err := config.New(config.DefaultOptions())
    if err != nil {
        panic(err)
    }

    // Initialize logger
    logger, err := logging.New(logging.DefaultConfig())
    if err != nil {
        panic(err)
    }

    // Use the components
    port := cfg.GetInt("server.port")
    logger.Info("starting server", zap.Int("port", port))
}
```

## Configuration

### Environment Variables

All configuration can be overridden using environment variables:

```bash
TERRAORBIT_DATABASE_HOST=localhost
TERRAORBIT_SERVER_PORT=8080
```

### Configuration File

Default configuration file (`config.yaml`):

```yaml
database:
  host: localhost
  port: 5432
  user: admin

server:
  port: 8080
  timeout: 30s
```

## Development

### Prerequisites

- Go 1.21+
- Make (optional)

### Testing

Run all tests:

```bash
go test ./...
```

Run specific package tests:

```bash
go test ./pkg/errors -v
go test ./pkg/logging -v
go test ./pkg/config -v
```

## Internal Contributing Guidelines

1. Create a feature branch (`git checkout -b feature/amazing-feature`)
2. Commit your changes (`git commit -m 'feat: Add some amazing feature'`)
3. Push to the branch (`git push origin feature/amazing-feature`)
4. Create a Pull Request

## License

This project is proprietary software owned by StackCatalyst. All rights reserved. See the LICENSE file for details.


Setting up Prometheus integration for advanced metrics collection
Implementing Jaeger tracing for distributed system monitoring
Creating an ELK Stack integration for centralized logging
Implementing AlertManager for advanced alerting