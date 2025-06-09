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
	news "itmrchow/tw-media-analytics-service/domain/news/delivery"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	"itmrchow/tw-media-analytics-service/domain/news/service"
	spider "itmrchow/tw-media-analytics-service/domain/spider/delivery"
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

	// ai model
	model := infra.InitAIModel(ctx, logger)

	// queue
	q := infra.InitQueue(ctx, logger)

	// cron
	jobs := cronjob.NewCronJob(logger, q)
	infra.InitCron(ctx, logger, jobs)

	// repository
	// TODO: init func
	_, repoSpan := tracer.Start(ctx, "main/main: Init Repository")
	newsRepo := repository.NewNewsRepositoryImpl(logger, db)
	authorRepo := repository.NewAuthorRepositoryImpl(logger, db)
	analysisRepo := repository.NewAnalysisRepositoryImpl(logger, db)
	repoSpan.End()

	// service
	// TODO: init func
	_, serviceSpan := tracer.Start(ctx, "main/main: Init Service")
	newsService := service.NewNewsServiceImpl(logger, newsRepo, authorRepo, analysisRepo, q, db, model)
	serviceSpan.End()

	// handler
	// - Spider handler
	// TODO: init func
	_, spiderSpan := tracer.Start(ctx, "main/main: Init Spider Consumer Handler")
	spiderEventHandlerMap := map[uint]*spider.SpiderEventHandler{
		1: spider.NewCtiNewsNewsSpiderEventHandler(logger, q), // 中天
		2: spider.NewSetnNewsSpiderEventHandler(logger, q),    // 三立
	}
	spiderEventHandler := spider.NewBaseEventHandler(logger, spiderEventHandlerMap)
	spiderSpan.End()

	// - news handler
	// TODO: init func
	_, newsSpan := tracer.Start(ctx, "main/main: Init News Consumer Handler")
	newsHandler := news.NewNewsEventHandler(logger, q, db, newsService)
	// news.NewNewsEventHandler(logger, q, db, newsService)
	newsSpan.End()

	// consumer
	consumerCtx, consumerSpan := tracer.Start(ctx, "main/main: Init  Consumer")

	// 初始化 Spider Consumer
	if err = spider.InitSpiderConsumer(consumerCtx, q, spiderEventHandler); err != nil {
		logger.Fatal().Err(err).Ctx(consumerCtx).Msg("failed to init spider consumer")
	}

	// 初始化 News Consumer
	if err = news.InitNewsConsumer(consumerCtx, q, newsHandler); err != nil {
		logger.Fatal().Err(err).Ctx(consumerCtx).Msg("failed to init news consumer")
	}

	consumerSpan.End()

	logger.Info().Ctx(ctx).Msg("server started ^^")
	span.End()

	defer func() {
		err = errors.Join(err, q.CloseClient())
		err = errors.Join(err, model.CloseClient())

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
