package logger

import (
	"context"
	"log/slog"
	"os"
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

// Handle extracts trace_id from ctx and prepends it to the record's attributes.
// If trace_id is non-empty, it adds it as the first attribute.
// It also removes any manually-added "trace_id" attribute to ensure context value takes precedence.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	traceID := GetTraceID(ctx)

	if traceID != "" {
		// Collect existing attributes, filtering out any manually-added "trace_id"
		attrs := make([]slog.Attr, 0, r.NumAttrs())
		r.Attrs(func(a slog.Attr) bool {
			if a.Key != "trace_id" {
				attrs = append(attrs, a)
			}
			return true
		})

		// Create a new record with trace_id prepended
		newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		newRecord.AddAttrs(slog.String("trace_id", traceID))
		newRecord.AddAttrs(attrs...)

		return h.inner.Handle(ctx, newRecord)
	}

	return h.inner.Handle(ctx, r)
}

// WithAttrs delegates to the inner handler, returning a new TraceHandler.
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup delegates to the inner handler, returning a new TraceHandler.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{inner: h.inner.WithGroup(name)}
}

// NewLogger creates a configured *slog.Logger with TraceHandler wrapping JSONHandler.
// Writes JSON to stdout at Info level with source information enabled.
func NewLogger() *slog.Logger {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})

	traceHandler := NewTraceHandler(jsonHandler)

	return slog.New(traceHandler)
}
