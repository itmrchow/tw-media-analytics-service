package logger

import (
	"bytes"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx/fxevent"
)

// TestFxZerologLogger_LogEvent 測試 FxZerologLogger 的 LogEvent 方法
func TestFxZerologLogger_LogEvent(t *testing.T) {
	// 準備測試用的 buffer 來捕獲日誌輸出
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel)
	fxLogger := NewFxZerologLogger(&logger)

	tests := []struct {
		name     string
		event    fxevent.Event
		expected string
	}{
		{
			name: "OnStartExecuting event",
			event: &fxevent.OnStartExecuting{
				FunctionName: "testFunction",
				CallerName:   "testCaller",
			},
			expected: "OnStart hook executing",
		},
		{
			name: "OnStartExecuted event with error",
			event: &fxevent.OnStartExecuted{
				FunctionName: "testFunction",
				CallerName:   "testCaller",
				Err:          assert.AnError,
			},
			expected: "OnStart hook failed",
		},
		{
			name: "OnStartExecuted event without error",
			event: &fxevent.OnStartExecuted{
				FunctionName: "testFunction",
				CallerName:   "testCaller",
				Runtime:      time.Millisecond * 100,
			},
			expected: "OnStart hook executed",
		},
		{
			name: "Provided event",
			event: &fxevent.Provided{
				ConstructorName: "testConstructor",
				OutputTypeNames: []string{"TestType"},
				ModuleName:      "testModule",
				Private:         false,
			},
			expected: "provided",
		},
		{
			name:     "Started event without error",
			event:    &fxevent.Started{},
			expected: "started",
		},
		{
			name: "Started event with error",
			event: &fxevent.Started{
				Err: assert.AnError,
			},
			expected: "start failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清空 buffer
			buf.Reset()

			// 執行測試
			fxLogger.LogEvent(tt.event)

			// 驗證輸出包含預期的訊息
			output := buf.String()
			assert.Contains(t, output, tt.expected, "日誌輸出應該包含預期的訊息")
		})
	}
}

// TestFxZerologLogger_NewFxZerologLogger 測試 NewFxZerologLogger 建構函數
func TestFxZerologLogger_NewFxZerologLogger(t *testing.T) {
	logger := zerolog.New(nil)
	fxLogger := NewFxZerologLogger(&logger)

	assert.NotNil(t, fxLogger, "FxZerologLogger 不應該為 nil")
	assert.Equal(t, &logger, fxLogger.logger, "logger 應該正確設置")
}

// TestFxZerologLogger_logEvent 測試 logEvent 方法
func TestFxZerologLogger_logEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.InfoLevel)
	fxLogger := NewFxZerologLogger(&logger)

	// 測試記錄一般事件
	fxLogger.logEvent("test message", "key1", "value1", "key2", "value2")

	output := buf.String()
	assert.Contains(t, output, "test message", "應該包含訊息")
	assert.Contains(t, output, "key1", "應該包含鍵")
	assert.Contains(t, output, "value1", "應該包含值")
}

// TestFxZerologLogger_logError 測試 logError 方法
func TestFxZerologLogger_logError(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.ErrorLevel)
	fxLogger := NewFxZerologLogger(&logger)

	// 測試記錄錯誤事件
	testErr := assert.AnError
	fxLogger.logError("test error", "error", testErr)

	output := buf.String()
	assert.Contains(t, output, "test error", "應該包含錯誤訊息")
	assert.Contains(t, output, "error", "應該包含錯誤欄位")
}
