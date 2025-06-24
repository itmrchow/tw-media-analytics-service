package otel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitOptel(
	ctx context.Context,
	logger *zerolog.Logger,
) (func(context.Context) error, error) {
	// 初始化 OpenTelemetry
	otelShutdown, err := SetupOTelSDK(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup otel sdk: %w", err)
	}

	return otelShutdown, err
}

func SetupOTelSDK(ctx context.Context, logger *zerolog.Logger) (func(context.Context) error, error) {
	logger.Info().Ctx(ctx).Msg("SetupOTelSDK: Setting up OpenTelemetry SDK")
	defer logger.Info().Ctx(ctx).Msg("SetupOTelSDK: OpenTelemetry SDK setup completed")

	var shutdownFuncs []func(context.Context) error
	var err error

	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		for _, fn := range shutdownFuncs {
			shutdownErr = errors.Join(shutdownErr, fn(ctx))
		}
		shutdownFuncs = nil
		return shutdownErr
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Resource 定義服務資訊
	res, err := newResource()
	if err != nil {
		handleErr(err)
		return nil, err
	}

	// Propagator
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Tracer Provider
	traceProvider, err := newTraceProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)

	// meter provider
	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	if viper.GetString("ENV") == "test" {
		return shutdown, nil
	}

	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(meterProvider)

	return shutdown, nil
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(viper.GetString("SERVICE_NAME")),
			semconv.ServiceVersion("v0.0.1"),
			semconv.DeploymentEnvironment(viper.GetString("ENV")),
		))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	// Create OTLP gRPC trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")), // Jaeger OTLP gRPC endpoint
		otlptracegrpc.WithInsecure(), // 開發環境使用非加密連線
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Duration(viper.GetInt("OTEL_BATCH_TIMEOUT"))*time.Second),
			trace.WithMaxExportBatchSize(viper.GetInt("OTEL_BATCH_SIZE")),
		),
		trace.WithResource(res),
	)

	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	// Create OTLP gRPC metric exporter
	// metricExporter, err := otlpmetricgrpc.New(ctx,
	// 	otlpmetricgrpc.WithEndpoint("localhost:4317"), // Jaeger OTLP gRPC endpoint
	// 	otlpmetricgrpc.WithInsecure(),                 // 開發環境使用非加密連線
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// meterProvider := metric.NewMeterProvider(
	// 	metric.WithReader(
	// 		metric.NewPeriodicReader(metricExporter,
	// 			metric.WithInterval(5*time.Second)),
	// 	),
	// 	metric.WithResource(res),
	// )

	meterProvider := metric.NewMeterProvider()

	return meterProvider, nil
}
