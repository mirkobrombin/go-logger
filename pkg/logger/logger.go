package logger

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the textual representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "unknown"
	}
}

// Field is a single structured key/value pair.
type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Entry is the log payload passed to sinks.
type Entry struct {
	Level  string                 `json:"level"`
	Time   time.Time              `json:"time"`
	Msg    string                 `json:"msg"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// Sink receives fully composed log entries.
type Sink interface {
	Log(e Entry) error
}

// Logger is the public logging contract.
type Logger interface {
	With(fields ...Field) Logger
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	RegisterSink(s Sink)
	SetLevel(l Level)
}

// Option configures the concrete logger on creation.
type Option func(*stdLogger)

// stdLogger is a simple, thread-safe structured logger with pluggable sinks.
type stdLogger struct {
	mu     sync.RWMutex
	sinks  []Sink
	level  Level
	fields map[string]interface{}
}

// New constructs a logger with optional options.
func New(opts ...Option) Logger {
	l := &stdLogger{
		level:  InfoLevel,
		fields: map[string]interface{}{},
		sinks:  []Sink{NewConsoleSink(nil)},
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// WithLevel sets the minimum level for emitted logs.
func WithLevel(level Level) Option {
	return func(l *stdLogger) { l.level = level }
}

// WithSink adds an initial sink.
func WithSink(s Sink) Option {
	return func(l *stdLogger) { l.sinks = append(l.sinks, s) }
}

// WithoutDefaultSink removes the default ConsoleSink added by New.
// Use before WithSink to create a logger with only custom sinks:
//
//	logger.New(logger.WithoutDefaultSink(), logger.WithSink(clefSink))
func WithoutDefaultSink() Option {
	return func(l *stdLogger) { l.sinks = nil }
}

// WithFields binds fields to the logger returned from New.
func WithFields(fields ...Field) Option {
	return func(l *stdLogger) {
		for _, f := range fields {
			l.fields[f.Key] = f.Value
		}
	}
}

// RegisterSink adds a sink at runtime.
func (l *stdLogger) RegisterSink(s Sink) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sinks = append(l.sinks, s)
}

// SetLevel changes the log level at runtime.
func (l *stdLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// With returns a derived logger that shares sinks but has extra bound fields.
func (l *stdLogger) With(fields ...Field) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	nextFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		nextFields[k] = v
	}
	for _, f := range fields {
		nextFields[f.Key] = f.Value
	}

	return &stdLogger{
		sinks:  l.sinks,
		level:  l.level,
		fields: nextFields,
	}
}

func (l *stdLogger) shouldLog(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.level
}

func (l *stdLogger) log(level Level, msg string, fields ...Field) {
	if !l.shouldLog(level) {
		return
	}

	entry := Entry{
		Level:  level.String(),
		Time:   time.Now().UTC(),
		Msg:    msg,
		Fields: map[string]interface{}{},
	}

	l.mu.RLock()
	for k, v := range l.fields {
		entry.Fields[k] = v
	}
	sinks := append([]Sink(nil), l.sinks...)
	l.mu.RUnlock()

	for _, f := range fields {
		entry.Fields[f.Key] = f.Value
	}

	for _, sink := range sinks {
		_ = sink.Log(entry)
	}
}

// Debug logs at debug level.
func (l *stdLogger) Debug(msg string, fields ...Field) { l.log(DebugLevel, msg, fields...) }

// Info logs at info level.
func (l *stdLogger) Info(msg string, fields ...Field) { l.log(InfoLevel, msg, fields...) }

// Warn logs at warn level.
func (l *stdLogger) Warn(msg string, fields ...Field) { l.log(WarnLevel, msg, fields...) }

// Error logs at error level.
func (l *stdLogger) Error(msg string, fields ...Field) { l.log(ErrorLevel, msg, fields...) }

// ConsoleSink writes entries as compact JSON lines to an io.Writer.
type ConsoleSink struct {
	w io.Writer
}

// NewConsoleSink constructs a ConsoleSink.
func NewConsoleSink(w io.Writer) *ConsoleSink {
	if w == nil {
		w = os.Stdout
	}
	return &ConsoleSink{w: w}
}

// Log writes a structured entry to the underlying writer.
func (c *ConsoleSink) Log(e Entry) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = c.w.Write(append(b, '\n'))
	return err
}
