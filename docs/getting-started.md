# Getting Started

`go-logger` keeps the main logging API intentionally small while allowing sinks to adapt entries to different outputs.

## Core pieces

- `logger.New` creates a logger with a JSON console sink by default.
- `logger.With` derives a logger with additional bound fields.
- `logger.RegisterSink` attaches new sinks at runtime.
- `logger.SetLevel` changes the minimum emitted log level.

## Included sinks

- `ConsoleSink` writes compact JSON lines.
- `RotatingFileSink` writes JSON lines to rolling log files.
- `PrometheusSink` tracks counts by level and exposes a scrape handler.
- `TelegramSink` forwards selected log entries to a Telegram chat.

## Operational note

The logger itself keeps sink delivery best-effort. Sinks still return errors so callers and tests can validate integration behavior when needed.

## Custom-sink-only logger

By default `logger.New` adds a `ConsoleSink` (JSON lines to stdout). When a service uses a different sink as its sole output — e.g. `CLEFSink` for Grafana Alloy — the default sink produces duplicate output in a different format.

Use `WithoutDefaultSink()` before `WithSink` to start from a clean slate:

```go
// CLEF-only output (e.g., for Grafana Alloy)
lg := logger.New(
    logger.WithoutDefaultSink(),
    logger.WithSink(logger.NewCLEFConsoleSink()),
)
```
