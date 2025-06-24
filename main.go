package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/ai"
	"itmrchow/tw-media-analytics-service/domain/cronjob"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	mAi "itmrchow/tw-media-analytics-service/domain/utils/ai"
	"itmrchow/tw-media-analytics-service/domain/utils/config"
	"itmrchow/tw-media-analytics-service/domain/utils/db"
	"itmrchow/tw-media-analytics-service/domain/utils/logger"
	"itmrchow/tw-media-analytics-service/domain/utils/mq"
	mOtel "itmrchow/tw-media-analytics-service/domain/utils/otel"
)

func main() {
	// 系統信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// config
	config.InitConfig()

	// context
	ctx, cancel := context.WithCancel(context.Background())

	// logger
	logger := logger.InitLogger()

	// fx
	app := fx.New(
		// Supply , 如果接Interface要用annotate註記
		fx.Supply(
			fx.Annotate(ctx, fx.As(new(context.Context)), fx.ResultTags(`name:"d_ctx"`)),
			fx.Annotate(cancel, fx.As(new(context.CancelFunc))),
			fx.Annotate(
				"tw-media-analytics-service", fx.ResultTags(`name:"tracer_name"`),
			),
			logger,
		),
		// OpenTelemetry
		fx.Provide(
			// mOtel.InitOptel,
			fx.Annotate(
				otel.Tracer,
				fx.As(new(trace.Tracer)),
				fx.ParamTags(`name:"tracer_name"`),
			),
			fx.Annotate(
				mOtel.InitOptel,
				fx.ParamTags(`name:"d_ctx"`),
			),
		),
		// Span Init Provider
		spanInitProvider(),
		// mq
		fx.Provide(
			fx.Annotate(
				mq.NewSubscriber,
				fx.As(new(message.Subscriber)),
			),
			fx.Annotate(
				mq.NewPublisher,
				fx.As(new(message.Publisher)),
			),
		),
		// db
		fx.Provide(
			db.NewMysqlDB,
		),
		// repository
		fx.Provide(
			repository.NewNewsRepositoryImpl,
			repository.NewAuthorRepositoryImpl,
			repository.NewAnalysisRepositoryImpl,
		),
		// ai
		fx.Provide(
			mAi.NewGemini,
		),
		// cronjob
		fx.Provide(
			cronjob.NewCronJob,
		),

		// // news module
		// fx.Provide(
		// 	fx.Annotate(
		// 		newsService.NewNewsServiceImpl,
		// 		fx.As(new(newsService.NewsService)),
		// 	),
		// 	fx.Annotate(
		// 		newsDelivery.NewNewsEventHandler,
		// 		fx.As(new(newsDelivery.NewsEventHandler)),
		// 	),
		// ),
		// // spider module
		// fx.Provide(
		// 	// Spider uc
		// 	fx.Annotate(
		// 		spiderUsecase.NewCtiNewsSpider,
		// 		fx.As(new(spiderUsecase.Spider)),
		// 		fx.ResultTags(`name:"cti_news_spider"`, `group:"spiders"`),
		// 	),
		// 	fx.Annotate(
		// 		spiderUsecase.NewSetnSpider,
		// 		fx.As(new(spiderUsecase.Spider)),
		// 		fx.ResultTags(`name:"setn_news_spider"`, `group:"spiders"`),
		// 	),
		// 	// Spider event handler
		// 	fx.Annotate(
		// 		spiderDelivery.NewCtiNewsNewsSpiderEventHandler,
		// 		fx.As(new(spiderDelivery.SpiderEventHandler)),
		// 		fx.ParamTags(`name:"cti_news_spider"`),
		// 		fx.ResultTags(`name:"cti_news_spider_event_handler"`, `group:"spider_event_handlers"`),
		// 	),
		// 	fx.Annotate(
		// 		spiderDelivery.NewSetnNewsSpiderEventHandler,
		// 		fx.As(new(spiderDelivery.SpiderEventHandler)),
		// 		fx.ParamTags(`name:"setn_news_spider"`),
		// 		fx.ResultTags(`name:"setn_news_spider_event_handler"`, `group:"spider_event_handlers"`),
		// 	),
		// 	fx.Annotate(
		// 		spiderDelivery.NewBaseEventHandler,
		// 		fx.ParamTags(`group:"spider_event_handlers"`),
		// 	),

		// 	// // Spider uc

		// 	// // Spider event handler
		// 	// spiderEventHandlerMap := map[uint]*spiderDelivery.SpiderEventHandler{
		// 	// 	1: spiderDelivery.NewCtiNewsNewsSpiderEventHandler(logger, spiderTracer, publisher, ctiNewsSpider), // 中天
		// 	// 	2: spiderDelivery.NewSetnNewsSpiderEventHandler(logger, spiderTracer, publisher, setnNewsSpider),   // 三立
		// 	// }
		// 	// spiderEventHandler := spiderDelivery.NewBaseEventHandler(logger, spiderTracer, spiderEventHandlerMap)
		// 	// spiderDelivery.InitSpiderSubscribe(ctx, logger, spiderTracer, subscriber, spiderEventHandler)

		fx.Invoke(
			// Otel register
			func(lf fx.Lifecycle, otelShutdown func(context.Context) error) {
				lf.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						logger.Info().Ctx(ctx).Msg("shutting down otel sdk")
						return otelShutdown(ctx)
					},
				})
			},
			// Ping DB
			db.PingDB,

			// subscribe init
			// - news subscribe
			// newsDelivery.InitNewsSubscribe,
			// - spider subscribe
			// spiderDelivery.InitSpiderSubscribe,

			// Init Cronjob
			cronjob.InitCronJob,
			// Span Init close
			func(logger *zerolog.Logger, ctx context.Context, span trace.Span) {
				logger.Info().Ctx(ctx).Msg("Init Server: end")
				span.End()
			},
			// LifeCycle manager
			func(lf fx.Lifecycle, logger *zerolog.Logger, aiModel ai.AiModel, ormDB *gorm.DB, subscriber message.Subscriber, publisher message.Publisher) {
				lf.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return connClose(ctx, logger, aiModel, ormDB, subscriber, publisher)
					},
				})
			},
		),

		// 	// fx.WithLogger(
		// 	// 	func(log *zerolog.Logger) fxevent.Logger {
		// 	// 		return logger.NewFxZerologLogger(log)
		// 	// 	},
		// 	// ),
		// ),
	)

	go func() {
		app.Run()
		logger.Info().Msg("server started ^^")
	}()

	select {
	case sig := <-sigChan:
		logger.Info().Msgf("收到系統信號: %v, 開始關閉服務", sig)
		cancel()
	case <-ctx.Done():
		logger.Info().Msg("服務開始關閉")
	}
}

// connClose 關閉所有連線的函數
// Args:
//
//	logger: 日誌記錄器
//	aiModel: AI 模型實例
//	ormDB: GORM 資料庫實例
//	subscriber: 訊息訂閱者
//	publisher: 訊息發布者
//
// Returns:
//
//	error: 關閉連線時的錯誤
func connClose(
	ctx context.Context,
	logger *zerolog.Logger,
	aiModel ai.AiModel,
	ormDB *gorm.DB,
	subscriber message.Subscriber,
	publisher message.Publisher,
) error {
	logger.Info().Ctx(ctx).Msg("Close Connection")

	var err error
	// Close AI Model
	err = errors.Join(err, aiModel.CloseClient())

	// Close DB
	sqlDB, dbErr := ormDB.DB()
	if dbErr == nil {
		err = errors.Join(err, sqlDB.Close())
	} else {
		err = errors.Join(err, dbErr)
	}

	// Close Subscriber
	err = errors.Join(err, subscriber.Close())

	// Close Publisher
	err = errors.Join(err, publisher.Close())

	if err != nil {
		logger.Error().Ctx(ctx).Err(err).Msg("Close Connection Failed")
	}

	return err
}

func spanInitProvider() fx.Option {
	return fx.Provide(
		fx.Annotate(
			func(ctx context.Context, logger *zerolog.Logger, tracer trace.Tracer) (context.Context, trace.Span) {
				ctx, span := tracer.Start(ctx, "main: Init Server")
				logger.Info().Ctx(ctx).Msg("Init Server: start")
				return ctx, span
			},
			fx.ParamTags(`name:"d_ctx"`),
		),
	)
}
