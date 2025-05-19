package infra

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
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
