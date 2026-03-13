# Getting Started

`go-logger` keeps the main logging API intentionally small while allowing sinks to adapt entries to different outputs.

## Core pieces

- `logger.New` creates a logger with a JSON console sink by default.
- `logger.With` derives a logger with additional bound fields.
- `logger.RegisterSink` attaches new sinks at runtime.
- `logger.SetLevel` changes the minimum emitted log level.

## Included sinks

- `ConsoleSink` writes compact JSON lines.
- `CLEFSink` writes Compact Log Event Format (CLEF) JSON lines, compatible with Serilog's `RenderedCompactJsonFormatter`. Useful when Go services share a log pipeline with C# services and field names must be consistent across both.
- `RotatingFileSink` writes JSON lines to rolling log files.
- `PrometheusSink` tracks counts by level and exposes a scrape handler.
- `TelegramSink` forwards selected log entries to a Telegram chat.

## CLEF sink

`CLEFSink` outputs one JSON object per line using the CLEF field conventions:

| Field | Description |
|-------|-------------|
| `@t`  | Timestamp in RFC3339Nano (UTC) |
| `@l`  | Level string — omitted for Information per the CLEF specification |
| `@m`  | Rendered message |
| _others_ | Structured fields flattened at root level |

Level mapping: `debug` → `Debug`, `warn` → `Warning`, `error` → `Error`. Information produces no `@l` field.

```go
sink := logger.NewCLEFSink(nil) // nil → os.Stdout
lg := logger.New(logger.WithSink(sink))
lg.Warn("disk space low", logger.Field{Key: "free_gb", Value: 2})
// {"@t":"2026-03-13T10:00:00Z","@l":"Warning","@m":"disk space low","free_gb":2}

lg.Info("request handled", logger.Field{Key: "status", Value: 200})
// {"@t":"2026-03-13T10:00:01Z","@m":"request handled","status":200}
```

## Operational note

The logger itself keeps sink delivery best-effort. Sinks still return errors so callers and tests can validate integration behavior when needed.
