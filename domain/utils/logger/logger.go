package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.uber.org/fx/fxevent"
)

// InitLogger 初始化 ZeroLog logger.
func InitLogger() *zerolog.Logger {
	var writer io.Writer
	var logLevel zerolog.Level

	if viper.GetString("ENV") == "local" || viper.GetString("ENV") == "test" {
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
		Logger().Hook(TracingHook{})

	return &logger
}

// FxZerologLogger 是用於fx的ZeroLog 的 logger 實現.
type FxZerologLogger struct {
	logger *zerolog.Logger
}

// NewFxZerologLogger 創建一個新的 FxZerologLogger 實例
// Args:
//
//	logger: ZeroLog logger 實例
//
// Returns:
//
//	*FxZerologLogger: 新的 logger 實例
func NewFxZerologLogger(logger *zerolog.Logger) *FxZerologLogger {
	return &FxZerologLogger{
		logger: logger,
	}
}

// LogEvent 記錄 fxevent 事件到 ZeroLog logger.
// Args:
//
//	event: fxevent 事件.
func (f *FxZerologLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		f.logEvent("OnStart hook executing",
			"callee", e.FunctionName,
			"caller", e.CallerName,
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			f.logError("OnStart hook failed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"error", e.Err,
			)
		} else {
			f.logEvent("OnStart hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.OnStopExecuting:
		f.logEvent("OnStop hook executing",
			"callee", e.FunctionName,
			"caller", e.CallerName,
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			f.logError("OnStop hook failed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"error", e.Err,
			)
		} else {
			f.logEvent("OnStop hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			f.logError("error encountered while applying options",
				"type", e.TypeName,
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
				"error", e.Err,
			)
		} else {
			f.logEvent("supplied",
				"type", e.TypeName,
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			f.logEvent("provided",
				"constructor", e.ConstructorName,
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
				"type", rtype,
				"private", e.Private,
			)
		}
		if e.Err != nil {
			f.logError("error encountered while applying options",
				"module", e.ModuleName,
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"error", e.Err,
			)
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			f.logEvent("replaced",
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			f.logError("error encountered while replacing",
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
				"error", e.Err,
			)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			f.logEvent("decorated",
				"decorator", e.DecoratorName,
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
				"type", rtype,
			)
		}
		if e.Err != nil {
			f.logError("error encountered while applying options",
				"stacktrace", e.StackTrace,
				"moduletrace", e.ModuleTrace,
				"module", e.ModuleName,
				"error", e.Err,
			)
		}
	case *fxevent.BeforeRun:
		f.logEvent("before run",
			"name", e.Name,
			"kind", e.Kind,
			"module", e.ModuleName,
		)
	case *fxevent.Run:
		if e.Err != nil {
			f.logError("error returned",
				"name", e.Name,
				"kind", e.Kind,
				"module", e.ModuleName,
				"error", e.Err,
			)
		} else {
			f.logEvent("run",
				"name", e.Name,
				"kind", e.Kind,
				"runtime", e.Runtime.String(),
				"module", e.ModuleName,
			)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		f.logEvent("invoking",
			"function", e.FunctionName,
			"module", e.ModuleName,
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			f.logError("invoke failed",
				"error", e.Err,
				"stack", e.Trace,
				"function", e.FunctionName,
				"module", e.ModuleName,
			)
		}
	case *fxevent.Stopping:
		f.logEvent("received signal",
			"signal", e.Signal.String(),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			f.logError("stop failed", "error", e.Err)
		}
	case *fxevent.RollingBack:
		f.logError("start failed, rolling back", "error", e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			f.logError("rollback failed", "error", e.Err)
		}
	case *fxevent.Started:
		if e.Err != nil {
			f.logError("start failed", "error", e.Err)
		} else {
			f.logEvent("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			f.logError("custom logger initialization failed", "error", e.Err)
		} else {
			f.logEvent("initialized custom fxevent.Logger", "function", e.ConstructorName)
		}
	}
}

// logEvent 記錄一般事件
// Args:
//
//	msg: 日誌訊息
//	fields: 鍵值對欄位
func (f *FxZerologLogger) logEvent(msg string, fields ...interface{}) {
	event := f.logger.Info()
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			value := fields[i+1]
			switch v := value.(type) {
			case string:
				event = event.Str(key, v)
			case []string:
				event = event.Strs(key, v)
			case bool:
				event = event.Bool(key, v)
			case error:
				event = event.Err(v)
			default:
				event = event.Interface(key, v)
			}
		}
	}
	event.Msg(msg)
}

// logError 記錄錯誤事件
// Args:
//
//	msg: 日誌訊息
//	fields: 鍵值對欄位
func (f *FxZerologLogger) logError(msg string, fields ...interface{}) {
	event := f.logger.Error()
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			value := fields[i+1]
			switch v := value.(type) {
			case string:
				event = event.Str(key, v)
			case []string:
				event = event.Strs(key, v)
			case bool:
				event = event.Bool(key, v)
			case error:
				event = event.Err(v)
			default:
				event = event.Interface(key, v)
			}
		}
	}
	event.Msg(msg)
}
