package infra

import (
	"context"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

func InitQueue(ctx context.Context, logger *zerolog.Logger) queue.Queue {
	// Trace
	ctx, span := tracer.Start(ctx, "infra/InitQueue: Init Queue")
	logger.Info().Ctx(ctx).Msg("InitQueue: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("InitQueue: end")
		span.End()
	}()

	// New PubSub
	q := queue.NewGcpPubSub(ctx, logger)

	// Init topic
	err := q.InitTopic()
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("InitQueue: failed to create topic")
	}

	return q
}
