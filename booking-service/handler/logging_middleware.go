package handler

import (
	"log/slog"
	"time"

	"booking/booking-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingMiddleware returns a Gin middleware that:
// 1. Generates UUID v4 as trace_id via uuid.New().String()
// 2. Stores trace_id in request context via logger.WithTraceID
// 3. Sets X-Trace-Id response header
// 4. Logs request completion with method, path, status, latency_ms, client_ip, trace_id
// 5. Uses slog.LogAttrs with slog.LevelError for status >= 500, slog.LevelInfo otherwise
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		traceID := uuid.New().String()

		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Trace-Id", traceID)

		c.Next()

		latencyMs := time.Since(start).Milliseconds()

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency_ms", latencyMs),
			slog.String("client_ip", c.ClientIP()),
			slog.String("trace_id", traceID),
		}

		if c.Writer.Status() >= 500 {
			slog.LogAttrs(ctx, slog.LevelError, "HTTP request completed", attrs...)
		} else {
			slog.LogAttrs(ctx, slog.LevelInfo, "HTTP request completed", attrs...)
		}
	}
}
