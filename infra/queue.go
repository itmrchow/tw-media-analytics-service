package infra

import (
	"context"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

func InitQueue(ctx context.Context, logger *zerolog.Logger) queue.Queue {
	// Trace
	tracer := getInfraTracer()
	ctx, span := tracer.Start(ctx, "InitQueue")
	logger.Info().Ctx(ctx).Msg("InitQueue: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitQueue: end")
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
