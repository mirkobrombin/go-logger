# Go Logger

> [!CAUTION]
> go-logger is now part of the [go-foundation](https://github.com/mirkobrombin/go-foundation) framework. The v1.0.0 release mirrors go-logger v2.0.0, but future versions may introduce breaking changes. Please migrate your project.

A small **structured logging** library for Go with pluggable sinks and predictable runtime behavior.

## Features

- **Structured Fields:** Attach contextual key/value pairs to every entry.
- **Pluggable Sinks:** Send logs to the console, rotating files, Prometheus-style counters, Telegram, or custom integrations.
- **Derived Loggers:** Bind shared fields with `With(...)` and reuse them across call sites.
- **Runtime Level Control:** Adjust log verbosity without rebuilding your application.

## Installation

```bash
go get github.com/mirkobrombin/go-logger
```

## Quick Start

```go
package main

import "github.com/mirkobrombin/go-logger/pkg/logger"

func main() {
    lg := logger.New()
    lg.Info("service started", logger.Field{Key: "port", Value: 8080})

    requestLogger := lg.With(logger.Field{Key: "request_id", Value: "abcd"})
    requestLogger.Info("handling request")
}
```

## Documentation

- [Getting Started](docs/getting-started.md)

## Rotating File Sink

```go
package main

import "github.com/mirkobrombin/go-logger/pkg/logger"

func main() {
    sink, err := logger.NewRotatingFileSink("var/log/my-app.log", logger.RotatingFileOptions{
        MaxSizeMB:  50,
        MaxBackups: 7,
        MaxAgeDays: 14,
        Compress:   true,
    })
    if err != nil {
        panic(err)
    }
    defer sink.Close()

    lg := logger.New(logger.WithSink(sink))
    lg.Info("file logging enabled")
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
