package infra

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

var (
	logger *zerolog.Logger
	tracer trace.Tracer
)

func SetInfraLogger(zLogger *zerolog.Logger) {
	logger = zLogger
}

func SetInfraTracer(zTracer trace.Tracer) {
	tracer = zTracer
}
