package logger_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mirkobrombin/go-logger/pkg/logger"
)

func TestConsoleSinkJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := logger.NewConsoleSink(buf)
	lg := logger.New(logger.WithSink(sink), logger.WithLevel(logger.DebugLevel))
	lg.Info("hello", logger.Field{Key: "k", Value: "v"})

	line, err := buf.ReadString('\n')
	if err != nil {
		t.Fatalf("ReadString() error = %v", err)
	}

	var entry logger.Entry
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if entry.Msg != "hello" {
		t.Fatalf("entry.Msg = %q, want %q", entry.Msg, "hello")
	}
	if entry.Level != "info" {
		t.Fatalf("entry.Level = %q, want %q", entry.Level, "info")
	}
	if got, ok := entry.Fields["k"]; !ok || got != "v" {
		t.Fatalf("entry.Fields = %v, want key k=v", entry.Fields)
	}
}

func TestLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := logger.NewConsoleSink(buf)
	lg := logger.New(logger.WithSink(sink), logger.WithLevel(logger.WarnLevel))
	lg.Info("should be filtered")
	if buf.Len() != 0 {
		t.Fatalf("buffer length = %d, want 0", buf.Len())
	}

	lg.Error("should appear")
	if buf.Len() == 0 {
		t.Fatalf("buffer length = 0, want > 0")
	}
}

func TestWithBindsContextFields(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := logger.NewConsoleSink(buf)
	lg := logger.New(logger.WithSink(sink), logger.WithFields(logger.Field{Key: "service", Value: "api"}))

	requestLogger := lg.With(logger.Field{Key: "request_id", Value: "abc"})
	requestLogger.Info("serving")

	line, err := buf.ReadString('\n')
	if err != nil {
		t.Fatalf("ReadString() error = %v", err)
	}

	var entry logger.Entry
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if entry.Fields["service"] != "api" {
		t.Fatalf("service field = %v, want api", entry.Fields["service"])
	}
	if entry.Fields["request_id"] != "abc" {
		t.Fatalf("request_id field = %v, want abc", entry.Fields["request_id"])
	}
}

func TestPrometheusSinkTracksCounts(t *testing.T) {
	sink := logger.NewPrometheusSink(logger.InfoLevel, "go_logger")
	if err := sink.Log(logger.Entry{Level: "debug", Time: time.Now(), Msg: "ignored"}); err != nil {
		t.Fatalf("Log() error = %v", err)
	}
	if err := sink.Log(logger.Entry{Level: "error", Time: time.Now(), Msg: "counted"}); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	sink.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, "debug") {
		t.Fatalf("metrics body unexpectedly contains debug counter: %q", body)
	}
	if !strings.Contains(body, "level=\"error\"") {
		t.Fatalf("metrics body = %q, want error counter", body)
	}
}

func TestNewTelegramSinkFromEnv(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_CHAT_ID", "chat")

	sink, err := logger.NewTelegramSinkFromEnv(logger.ErrorLevel)
	if err != nil {
		t.Fatalf("NewTelegramSinkFromEnv() error = %v", err)
	}
	if sink == nil {
		t.Fatalf("NewTelegramSinkFromEnv() = nil, want non-nil")
	}
}

func TestRotatingFileSinkWritesJSONLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	sink, err := logger.NewRotatingFileSink(path, logger.RotatingFileOptions{MaxSizeMB: 1})
	if err != nil {
		t.Fatalf("NewRotatingFileSink() error = %v", err)
	}
	defer sink.Close()

	entry := logger.Entry{
		Level: "info",
		Time:  time.Now().UTC(),
		Msg:   "written-to-file",
		Fields: map[string]interface{}{
			"service": "api",
		},
	}
	if err := sink.Log(entry); err != nil {
		t.Fatalf("Log() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "\"msg\":\"written-to-file\"") {
		t.Fatalf("log file content = %q, want serialized entry", string(content))
	}
}

func TestRotatingFileSinkRotates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	sink, err := logger.NewRotatingFileSink(path, logger.RotatingFileOptions{MaxSizeMB: 1, MaxBackups: 2})
	if err != nil {
		t.Fatalf("NewRotatingFileSink() error = %v", err)
	}
	defer sink.Close()

	largeMessage := strings.Repeat("x", 600*1024)
	for i := 0; i < 3; i++ {
		if err := sink.Log(logger.Entry{Level: "info", Time: time.Now().UTC(), Msg: largeMessage}); err != nil {
			t.Fatalf("Log() iteration %d error = %v", i, err)
		}
	}

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), "app*.log*"))
	if err != nil {
		t.Fatalf("filepath.Glob() error = %v", err)
	}
	if len(matches) < 2 {
		t.Fatalf("rotated files = %v, want at least 2 files", matches)
	}
}

func TestWithoutDefaultSink(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := logger.NewConsoleSink(buf)
	lg := logger.New(logger.WithoutDefaultSink(), logger.WithSink(sink))
	lg.Info("hello")
	if buf.Len() == 0 {
		t.Fatal("expected custom sink to receive output, got empty buffer")
	}
}
