package logger

import (
	"encoding/json"
	"errors"

	"gopkg.in/natefinch/lumberjack.v2"
)

// RotatingFileOptions configures file rotation behavior.
type RotatingFileOptions struct {
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	LocalTime  bool
}

// RotatingFileSink writes JSON log lines to a rolling file.
type RotatingFileSink struct {
	writer *lumberjack.Logger
}

// NewRotatingFileSink creates a sink backed by a rolling log file.
func NewRotatingFileSink(path string, opts RotatingFileOptions) (*RotatingFileSink, error) {
	if path == "" {
		return nil, errors.New("logger: rotating file path is required")
	}

	if opts.MaxSizeMB <= 0 {
		opts.MaxSizeMB = 10
	}

	return &RotatingFileSink{
		writer: &lumberjack.Logger{
			Filename:   path,
			MaxSize:    opts.MaxSizeMB,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAgeDays,
			Compress:   opts.Compress,
			LocalTime:  opts.LocalTime,
		},
	}, nil
}

// Log writes a structured entry to the rolling log file.
func (r *RotatingFileSink) Log(e Entry) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = r.writer.Write(append(payload, '\n'))
	return err
}

// Rotate forces the underlying writer to rotate the active log file.
func (r *RotatingFileSink) Rotate() error {
	return r.writer.Rotate()
}

// Close closes the underlying rotating log writer.
func (r *RotatingFileSink) Close() error {
	return r.writer.Close()
}
