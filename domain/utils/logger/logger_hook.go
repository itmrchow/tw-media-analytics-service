package logger

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// TracingHook 是 ZeroLog 的 hook，用於記錄 OpenTelemetry 的 trace ID 和 span ID.
type TracingHook struct{}

func (h TracingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {

	ctx := e.GetCtx()
	span := trace.SpanFromContext(ctx)

	if span == nil {
		return
	}
	spanCtx := span.SpanContext()
	traceID := spanCtx.TraceID()
	spanID := spanCtx.SpanID()

	if traceID.IsValid() {
		e.Str("trace_id", traceID.String())
	}
	if spanID.IsValid() {
		e.Str("span_id", spanID.String())
	}

	if level >= zerolog.InfoLevel {
		span.AddEvent(msg)
	}
}
