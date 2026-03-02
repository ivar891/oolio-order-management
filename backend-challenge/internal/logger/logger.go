package logger

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceIDContextKey is the key used to store the TraceID in the log record.
const TraceIDContextKey = "trace_id"

// New creates a new structured logger based on the environment.
func New(env string, out io.Writer) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "production" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	// Wrap handler with OTel trace ID extractor
	handler = &otelslogHandler{handler}

	return slog.New(handler)
}

// otelslogHandler is a middleware handler that extracts TraceID from context.
type otelslogHandler struct {
	slog.Handler
}

func (h *otelslogHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.Handler.Handle(ctx, r)
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(slog.String(TraceIDContextKey, span.SpanContext().TraceID().String()))
	}

	return h.Handler.Handle(ctx, r)
}

// WithGroup returns a new handler with the given group name.
func (h *otelslogHandler) WithGroup(name string) slog.Handler {
	return &otelslogHandler{h.Handler.WithGroup(name)}
}

// WithAttrs returns a new handler with the given attributes.
func (h *otelslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelslogHandler{h.Handler.WithAttrs(attrs)}
}

// Global sets the global slog logger.
func SetGlobal(l *slog.Logger) {
	slog.SetDefault(l)
}
