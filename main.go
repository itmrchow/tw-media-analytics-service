package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/cronjob"
	newsDelivery "itmrchow/tw-media-analytics-service/domain/news/delivery"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	nService "itmrchow/tw-media-analytics-service/domain/news/service"
	spiderDelivery "itmrchow/tw-media-analytics-service/domain/spider/delivery"
	spiderUsecase "itmrchow/tw-media-analytics-service/domain/spider/usecase"
	"itmrchow/tw-media-analytics-service/infra"
)

func main() {
	// 系統信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// error
	var err error

	// config
	infra.InitConfig()

	// logger
	logger := infra.InitLogger()

	// context
	ctx, cancel := context.WithCancel(context.Background())

	// Set up OpenTelemetry
	otelShutdown, err := infra.SetupOTelSDK(ctx, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to setup otel sdk")
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	tracer := otel.Tracer("tw-media-analytics-service")

	infra.SetInfraLogger(logger)
	infra.SetInfraTracer(tracer)

	// SpanName format: [pkg]/[func]: [description]
	ctx, span := tracer.Start(ctx, "main/main: Init Server", trace.WithAttributes(
	// attribute.String("operation.type", "infra_init"),
	))

	// db
	db := infra.InitMysqlDB(ctx)

	// mq:subscriber
	subscriber := infra.InitSubscriber(ctx, logger)

	// mq:publisher
	publisher := infra.InitPublisher(ctx, logger)

	// repository
	_, repoSpan := tracer.Start(ctx, "main/main: Init Repository")
	newsRepo := repository.NewNewsRepositoryImpl(logger, db)
	authorRepo := repository.NewAuthorRepositoryImpl(logger, db)
	analysisRepo := repository.NewAnalysisRepositoryImpl(logger, db)
	repoSpan.End()

	// Module
	// Init AI Model
	aiModel := infra.InitAIModel(ctx, logger)
	// Init News
	newsTracer := otel.Tracer("tw-media-analytics-service:news")
	// TODO: init func
	_, newsSpan := newsTracer.Start(ctx, "main/main: Init News Module")
	newsService := nService.NewNewsServiceImpl(
		logger,
		newsTracer,
		newsRepo,
		authorRepo,
		analysisRepo,
		publisher,
		db,
		aiModel,
	)
	newsHandler := newsDelivery.NewNewsEventHandler(logger, newsTracer, db, newsService)
	newsDelivery.InitNewsSubscribe(ctx, logger, newsTracer, subscriber, newsHandler)
	newsSpan.End()

	// Init Spider
	spiderTracer := otel.Tracer("tw-media-analytics-service:spider")
	_, spiderSpan := spiderTracer.Start(ctx, "main/main: Init Spider Module")
	// Spider uc
	ctiNewsSpider := spiderUsecase.NewCtiNewsSpider(logger, spiderTracer)
	setnNewsSpider := spiderUsecase.NewSetnSpider(logger, spiderTracer)
	// Spider event handler
	spiderEventHandlerMap := map[uint]*spiderDelivery.SpiderEventHandler{
		1: spiderDelivery.NewCtiNewsNewsSpiderEventHandler(logger, spiderTracer, publisher, ctiNewsSpider), // 中天
		2: spiderDelivery.NewSetnNewsSpiderEventHandler(logger, spiderTracer, publisher, setnNewsSpider),   // 三立
	}
	spiderEventHandler := spiderDelivery.NewBaseEventHandler(logger, spiderTracer, spiderEventHandlerMap)
	spiderDelivery.InitSpiderSubscribe(ctx, logger, spiderTracer, subscriber, spiderEventHandler)
	spiderSpan.End()

	// Init Cron
	cronTracer := otel.Tracer("tw-media-analytics-service:cron")
	cronJob := cronjob.NewCronJob(logger, cronTracer, publisher)
	cronJob.InitCron(ctx)

	logger.Info().Ctx(ctx).Msg("server started ^^")
	span.End()

	defer func() {
		err = errors.Join(err, aiModel.CloseClient())
		err = errors.Join(err, subscriber.Close())
		err = errors.Join(err, publisher.Close())

		logger.Info().Ctx(ctx).Msg("Client closed")
	}()

	select {
	case sig := <-sigChan:
		logger.Info().Msgf("收到系統信號: %v, 開始關閉服務", sig)
		cancel()
	case <-ctx.Done():
		logger.Info().Msg("服務開始關閉")
	}

}
