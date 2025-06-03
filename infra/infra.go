package infra

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	logger *zerolog.Logger
)

func SetInfraLogger(zLogger *zerolog.Logger) {
	logger = zLogger
}

// getInfraTracer 取得 infra 的 tracer.
func getInfraTracer() trace.Tracer {
	return otel.Tracer("infra")
}
