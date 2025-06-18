package delivery

import (
	"context"

	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

// InitNewsSubscribe 初始化新聞相關訂閱.
func InitNewsSubscribe(
	ctx context.Context,
	logger *zerolog.Logger,
	tracer trace.Tracer,
	subscriber *googlecloud.Subscriber,
	handler *NewsEventHandler,
) {
	// Tracer
	ctx, span := tracer.Start(ctx, "domain/news/delivery/InitNewsSubscribe: Init News Subscribe")
	logger.Info().Ctx(ctx).Msg("InitNewsSubscribe: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("InitNewsSubscribe: end")
		span.End()
	}()

	// subscribe
	group, ctx := errgroup.WithContext(ctx)

	// - NewsCheck
	group.Go(func() error {
		newsCheckMsg, err := subscriber.Subscribe(ctx, string(queue.TopicNewsCheck))
		if err != nil {
			logger.Error().Ctx(ctx).Err(err).Msg("failed to subscribe news check")
			return err
		}
		go process(logger, newsCheckMsg, handler.CheckNewsExistHandle)
		return nil
	})

	// - NewsSave
	group.Go(func() error {
		newsSaveMsg, err := subscriber.Subscribe(ctx, string(queue.TopicNewsSave))
		if err != nil {
			logger.Error().Ctx(ctx).Err(err).Msg("failed to subscribe news save")
			return err
		}
		go process(logger, newsSaveMsg, handler.SaveNewsHandle)
		return nil
	})

	// - GetAnalysis
	group.Go(func() error {
		getAnalysisMsg, err := subscriber.Subscribe(ctx, string(queue.TopicGetAnalysis))
		if err != nil {
			logger.Error().Ctx(ctx).Err(err).Msg("failed to subscribe get analysis")
			return err
		}
		go process(logger, getAnalysisMsg, handler.GetAnalysisHandle)
		return nil
	})
}

func process(
	logger *zerolog.Logger,
	messages <-chan *message.Message,
	handler func(ctx context.Context, msg []byte) error,
) {
	for msg := range messages {
		// call handler
		if err := handler(msg.Context(), msg.Payload); err != nil {
			logger.Error().Ctx(msg.Context()).Err(err).Bytes("payload", msg.Payload).Msg("failed to process message")
			return
		}
		msg.Ack()
	}
}
