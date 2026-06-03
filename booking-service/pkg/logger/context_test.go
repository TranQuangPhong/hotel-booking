package logger

import (
	"context"
	"testing"
)

func TestWithTraceID_StoresValue(t *testing.T) {
	ctx := context.Background()
	traceID := "abc-123-def"
	ctx = WithTraceID(ctx, traceID)

	got := GetTraceID(ctx)
	if got != traceID {
		t.Errorf("GetTraceID() = %q, want %q", got, traceID)
	}
}

func TestGetTraceID_ReturnsEmptyWhenMissing(t *testing.T) {
	ctx := context.Background()
	got := GetTraceID(ctx)
	if got != "" {
		t.Errorf("GetTraceID() = %q, want empty string", got)
	}
}

func TestGetTraceID_ReturnsEmptyOnNilContext(t *testing.T) {
	got := GetTraceID(nil)
	if got != "" {
		t.Errorf("GetTraceID(nil) = %q, want empty string", got)
	}
}

func TestWithTraceID_EmptyString(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "")

	got := GetTraceID(ctx)
	if got != "" {
		t.Errorf("GetTraceID() = %q, want empty string", got)
	}
}

func TestWithTraceID_OverridesPreviousValue(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "first")
	ctx = WithTraceID(ctx, "second")

	got := GetTraceID(ctx)
	if got != "second" {
		t.Errorf("GetTraceID() = %q, want %q", got, "second")
	}
}
