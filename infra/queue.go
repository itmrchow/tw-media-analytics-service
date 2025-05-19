package infra

import (
	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

func InitQueue(logger *zerolog.Logger, q queue.Queue) {
	// init topic
	err := q.InitTopic()
	if err == nil {
		logger.Info().Msg("Queue topic created")
	} else {
		logger.Fatal().Err(err).Msg("failed to create topic")
	}
}
