package infra

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	logI "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
)

func InitLogger() *zerolog.Logger {
	var writer io.Writer
	var logLevel zerolog.Level

	if viper.GetString("ENV") == "local" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.DateTime,
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("message=%s", i)
			},
			// 設定為 true 會讓 JSON 格式化輸出
			NoColor: false, // 設定為 true 會關閉顏色
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
		}
		logLevel = zerolog.DebugLevel
	} else {
		writer = os.Stdout
		logLevel = zerolog.InfoLevel
	}

	logger := zerolog.New(writer).Level(logLevel)
	logger = logger.With().
		Str("service", "tw-media-analytics-service").
		Time("time", time.Now()).
		Caller().
		Logger()
	return &logger
}

var _ logI.LoggerProvider = (*LoggerProvider)(nil)

type LoggerProvider struct {
	*log.LoggerProvider
}

func (l *LoggerProvider) Logger(name string, options ...logI.LoggerOption) logI.Logger {
	panic("TODO: Implement")
}

// func (l *LoggerProvider) loggerProvider() {
// 	panic("TODO: Implement")
// }

func NewLoggerProvider(opts ...log.LoggerProviderOption) *LoggerProvider {
	provider := log.NewLoggerProvider(opts...)
	return &LoggerProvider{
		LoggerProvider: provider,
	}
}

func (p *LoggerProvider) Shutdown(ctx context.Context) error {
	return p.LoggerProvider.Shutdown(ctx)
}

// Logger returns an OpenTelemetry Logger that does not record any telemetry.
// func (p *LoggerProvider) Logger(name string, opts ...log.LoggerOption) log.Logger {
// 	return Logger{}
// }

// type Logger struct {
// 	noop.Logger

// 	name       string
// 	version    string
// 	schemaURL  string
// 	attributes map[string]interface{}
// 	provider   *LoggerProvider
// }

// func (Logger) Emit(context.Context, log.Record) {}

// // Enabled returns false. No log records are ever emitted.
// func (Logger) Enabled(context.Context, log.EnabledParameters) bool { return false }
