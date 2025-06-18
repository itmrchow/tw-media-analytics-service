package infra

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

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
	err := q.InitTopic(ctx)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("InitQueue: failed to create topic")
	}

	return q
}

// InitSubscriber 初始化 subscriber , 用於設定訂閱.
func InitSubscriber(ctx context.Context, logger *zerolog.Logger) *googlecloud.Subscriber {
	// Tracer
	ctx, span := tracer.Start(ctx, "infra/InitSubscriber: Init Subscriber")
	logger.Info().Ctx(ctx).Msg("InitSubscriber: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("InitSubscriber: end")
		span.End()
	}()

	// TODO: 修改logger
	watermillLogger := watermill.NewStdLogger(false, false)

	subscriber, err := googlecloud.NewSubscriber(
		googlecloud.SubscriberConfig{
			GenerateSubscriptionName: func(topic string) string {
				return topic + "_" + viper.GetString("ENV") + "_sub"
			},
			ProjectID: viper.GetString("GCP_PROJECT_ID"),
			ClientConfig: &pubsub.ClientConfig{
				EnableOpenTelemetryTracing: true,
			},
		},
		watermillLogger,
	)
	if err != nil {
		logger.Fatal().Ctx(ctx).Err(err).Msg("InitSubscriber: failed to create subscriber")
	}

	return subscriber
}

// InitPublisher 初始化 publisher , 用於發送訊息.
func InitPublisher(ctx context.Context, logger *zerolog.Logger) *googlecloud.Publisher {
	// Tracer
	ctx, span := tracer.Start(ctx, "infra/InitPublisher: Init Publisher")
	logger.Info().Ctx(ctx).Msg("InitPublisher: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("InitPublisher: end")
		span.End()
	}()

	// TODO: 修改logger
	watermillLogger := watermill.NewStdLogger(false, false)

	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: viper.GetString("GCP_PROJECT_ID"),
	}, watermillLogger)
	if err != nil {
		logger.Fatal().Ctx(ctx).Err(err).Msg("InitPublisher: failed to create publisher")
	}

	return publisher
}
