# TerraOrbit Common Library Integration Guide

## Installation

Add the common library to your Go module:

```bash
go get github.com/StackCatalyst/common-lib
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/StackCatalyst/common-lib/pkg/config"
    "github.com/StackCatalyst/common-lib/pkg/logging"
    "github.com/StackCatalyst/common-lib/pkg/metrics"
    "github.com/StackCatalyst/common-lib/pkg/tracing"
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

    // Initialize metrics collector
    metricsCollector, err := metrics.NewCollector(metrics.DefaultCollectorConfig())
    if err != nil {
        panic(err)
    }

    // Initialize tracer
    tracer, err := tracing.New(tracing.DefaultConfig())
    if err != nil {
        panic(err)
    }
    defer tracer.Close()

    // Start metrics collection
    ctx := context.Background()
    if err := metricsCollector.Start(ctx); err != nil {
        logger.Error("Failed to start metrics collector", err)
    }
    defer metricsCollector.Stop(ctx)

    // Your service code here
}
```

## Core Components

### 1. Configuration Management

```go
import "github.com/StackCatalyst/common-lib/pkg/config"

// Load configuration from file
cfg, err := config.New(config.Options{
    ConfigFile: "config.yaml",
    EnvPrefix: "APP",
})

// Access configuration
dbHost := cfg.GetString("database.host")
serverPort := cfg.GetInt("server.port")
```

### 2. Structured Logging

```go
import "github.com/StackCatalyst/common-lib/pkg/logging"

// Initialize logger with rotation
logger, err := logging.NewAdvanced(logging.AdvancedConfig{
    Level:      "info",
    Format:     "json",
    OutputPath: "logs/app.log",
    MaxSize:    100, // MB
    MaxBackups: 5,
    MaxAge:     30, // days
})

// Contextual logging
logger.Info("Request processed",
    "method", "POST",
    "path", "/api/v1/users",
    "duration", 235,
)

// Error logging
if err != nil {
    logger.Error("Failed to process request", err,
        "user_id", userID,
    )
}
```

### 3. Metrics Collection

```go
import "github.com/StackCatalyst/common-lib/pkg/metrics"

// Initialize with standard metrics
reporter := metrics.NewStandardReporter(metrics.Options{
    Namespace: "myapp",
    Subsystem: "api",
}, metrics.StandardLabels{
    Service:     "user-service",
    Environment: "production",
})

// Use predefined metrics
httpReqs := reporter.Counter(
    metrics.MetricHTTPRequestsTotal,
    "Total HTTP requests",
    []string{"method", "path"},
)

// Record metrics
httpReqs.WithLabelValues("POST", "/api/users").Inc()
```

### 4. Distributed Tracing

```go
import "github.com/StackCatalyst/common-lib/pkg/tracing"

// Initialize tracer
tracer, err := tracing.New(tracing.Config{
    ServiceName: "user-service",
    AgentHost:   "jaeger-agent",
    AgentPort:   "6831",
})

// Create spans
span := tracer.StartSpan("process-request")
defer span.Finish()

// Add context
ctx := opentracing.ContextWithSpan(context.Background(), span)

// Add tags
tracing.WithField(ctx, "user_id", userID)
```

## HTTP Server Integration

```go
import (
    "github.com/StackCatalyst/common-lib/pkg/logging"
    "github.com/StackCatalyst/common-lib/pkg/metrics"
    "github.com/StackCatalyst/common-lib/pkg/tracing"
)

func setupServer(logger *logging.Logger, metrics *metrics.Reporter, tracer *tracing.Tracer) *http.Server {
    // Create middleware chain
    middleware := []func(http.Handler) http.Handler{
        logging.HTTPMiddleware(logger),
        metrics.HTTPMiddleware(),
        tracing.HTTPMiddleware(tracer),
    }

    // Apply middleware
    handler := yourHandler()
    for _, m := range middleware {
        handler = m(handler)
    }

    return &http.Server{
        Handler: handler,
    }
}
```

## gRPC Server Integration

```go
import (
    "github.com/StackCatalyst/common-lib/pkg/logging"
    "github.com/StackCatalyst/common-lib/pkg/metrics"
    "github.com/StackCatalyst/common-lib/pkg/tracing"
)

func setupGRPCServer(logger *logging.Logger, metrics *metrics.Reporter, tracer *tracing.Tracer) *grpc.Server {
    // Create interceptors
    interceptors := []grpc.UnaryServerInterceptor{
        logging.UnaryServerInterceptor(logger),
        metrics.UnaryServerInterceptor(),
        tracing.UnaryServerInterceptor(tracer),
    }

    // Chain interceptors
    chain := grpc.ChainUnaryInterceptor(interceptors...)

    return grpc.NewServer(chain)
}
```

## Error Handling

```go
import "github.com/StackCatalyst/common-lib/pkg/errors"

// Create errors
err := errors.New(errors.ErrNotFound, "user not found")

// Wrap errors
wrappedErr := errors.Wrap(err, errors.ErrInternal,
    "failed to process request")

// Check error types
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found error
}

// Use error groups for batch operations
group := errors.NewErrorGroup()
for _, item := range items {
    if err := process(item); err != nil {
        group.Add(err)
    }
}

if group.HasErrors() {
    logger.Error("Batch processing failed", group)
}
```

## Best Practices

1. **Configuration**
   - Use environment variables for sensitive values
   - Keep configuration files in version control
   - Use different configurations per environment

2. **Logging**
   - Use structured logging with consistent fields
   - Include request IDs for tracing
   - Set appropriate log levels

3. **Metrics**
   - Use standard metric names and labels
   - Monitor both business and technical metrics
   - Set up alerts for critical thresholds

4. **Tracing**
   - Propagate trace context across service boundaries
   - Add meaningful span tags
   - Use appropriate sampling rates

## Kubernetes Integration

See [Kubernetes Setup Guide](k8s/README.md) for:
- Service configuration
- Metrics scraping setup
- Tracing agent configuration
- Log aggregation

## Support

For issues and feature requests:
- GitHub Issues: [github.com/StackCatalyst/common-lib/issues](https://github.com/StackCatalyst/common-lib/issues)
- Internal Documentation: [confluence/common-lib](https://stackcatalyst.atlassian.net/wiki/common-lib) 