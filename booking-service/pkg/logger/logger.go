package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// TraceHandler wraps an slog.Handler to automatically inject trace_id from context.
type TraceHandler struct {
	inner slog.Handler
}

// NewTraceHandler creates a TraceHandler wrapping the given handler.
func NewTraceHandler(inner slog.Handler) *TraceHandler {
	return &TraceHandler{inner: inner}
}

// Enabled delegates to the inner handler.
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle extracts trace_id from ctx and prepends it as the first attribute.
// If trace_id is non-empty, it adds it as the first attribute.
// It also removes any manually-added "trace_id" attribute to ensure context value takes precedence.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID := GetTraceID(ctx)
	if traceID != "" {
		attrs := make([]slog.Attr, 0, r.NumAttrs())
		r.Attrs(func(a slog.Attr) bool {
			if a.Key != "trace_id" {
				attrs = append(attrs, a)
			}
			return true
		})
		newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		newRecord.AddAttrs(slog.String("trace_id", traceID))
		newRecord.AddAttrs(attrs...)
		return h.inner.Handle(ctx, newRecord)
	}
	return h.inner.Handle(ctx, r)
}

// WithAttrs returns a new TraceHandler wrapping inner.WithAttrs(attrs).
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a new TraceHandler wrapping inner.WithGroup(name).
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{inner: h.inner.WithGroup(name)}
}

// NewLogger creates a configured *slog.Logger with TraceHandler wrapping JSONHandler.
// Reads LOG_LEVEL env var (Debug, Info, Warn, Error — case-insensitive).
// Falls back to slog.LevelInfo when missing or invalid.
// Emits a warning log if the value is present but invalid.
func NewLogger() *slog.Logger {
	level := slog.LevelInfo
	invalidLevel := ""

	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		switch strings.ToLower(envLevel) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			invalidLevel = envLevel
		}
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})

	traceHandler := NewTraceHandler(jsonHandler)
	logger := slog.New(traceHandler)

	if invalidLevel != "" {
		fmt.Fprintf(os.Stderr, "WARNING: invalid LOG_LEVEL %q, falling back to Info\n", invalidLevel)
		logger.Warn("invalid LOG_LEVEL value, falling back to Info",
			slog.String("LOG_LEVEL", invalidLevel),
		)
	}

	return logger
}
