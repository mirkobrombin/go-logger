package logger

import (
	"fmt"
	"net/http"
	"sync"
)

// PrometheusSink exposes simple counters grouped by log level.
type PrometheusSink struct {
	mu       sync.Mutex
	counters map[string]int64
	minLevel Level
}

// NewPrometheusSink creates a Prometheus-compatible metrics sink.
func NewPrometheusSink(minLevel Level, namespace string) *PrometheusSink {
	_ = namespace
	return &PrometheusSink{counters: make(map[string]int64), minLevel: minLevel}
}

// Handler returns an HTTP handler serving a basic Prometheus text exposition.
func (p *PrometheusSink) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.mu.Lock()
		defer p.mu.Unlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		for level, value := range p.counters {
			_, _ = fmt.Fprintf(w, "go_logger_logs_total{level=\"%s\"} %d\n", level, value)
		}
	})
}

// Log increments the counter for the provided entry level when it passes the threshold.
func (p *PrometheusSink) Log(e Entry) error {
	if levelFromString(e.Level) < p.minLevel {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.counters[e.Level]++
	return nil
}
