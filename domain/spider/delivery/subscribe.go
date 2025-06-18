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

// InitSpiderSubscribe 初始化爬蟲相關訂閱.
func InitSpiderSubscribe(
	ctx context.Context,
	logger *zerolog.Logger,
	tracer trace.Tracer,
	subscriber *googlecloud.Subscriber,
	handler *BaseEventHandler,
) {
	// Tracer
	ctx, span := tracer.Start(ctx, "domain/spider/delivery/InitSpiderSubscribe: Init Spider Subscribe")
	logger.Info().Ctx(ctx).Msg("InitSpiderSubscribe: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("InitSpiderSubscribe: end")
		span.End()
	}()

	// subscribe
	group, ctx := errgroup.WithContext(ctx)

	// subscribe
	// - ArticleListScraping
	for mediaID, spiderHandler := range handler.SpiderMap {
		group.Go(func() error {
			logger.Debug().Ctx(ctx).Msgf("subscribe article list scraping: %d", mediaID)

			articleListScrapingMsg, err := subscriber.Subscribe(
				context.Background(),
				string(queue.TopicArticleListScraping),
			)
			if err != nil {
				logger.Error().Ctx(ctx).Err(err).Msg("failed to subscribe article content scraping")
				return err
			}

			go process(logger, articleListScrapingMsg, spiderHandler.ArticleListScrapingHandle)

			return nil
		})
	}

	// - ArticleContentScraping
	group.Go(func() error {
		articleContentScrapingMsg, err := subscriber.Subscribe(
			context.Background(),
			string(queue.TopicArticleContentScraping),
		)
		if err != nil {
			logger.Error().Ctx(ctx).Err(err).Msg("failed to subscribe article content scraping")
			return err
		}

		go process(logger, articleContentScrapingMsg, handler.ArticleContentScrapingHandle)

		return nil
	})
}

// TODO: 移到 utils
func process(
	logger *zerolog.Logger,
	messages <-chan *message.Message,
	handler func(ctx context.Context, msg []byte) error,
) {
	for msg := range messages {
		if err := handler(msg.Context(), msg.Payload); err != nil {
			logger.Error().Ctx(msg.Context()).Err(err).Bytes("payload", msg.Payload).Msg("failed to process message")
			return
		}
		msg.Ack()
	}
}
