package infra

import (
	"context"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

func InitQueue(ctx context.Context, logger *zerolog.Logger) queue.Queue {
	// create q obj
	q := queue.NewGcpPubSub(ctx, logger)

	// init topic
	err := q.InitTopic()
	if err == nil {
		logger.Info().Msg("Queue topic created")
	} else {
		logger.Fatal().Err(err).Msg("failed to create topic")
	}

	return q
}
