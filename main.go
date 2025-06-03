package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	"itmrchow/tw-media-analytics-service/domain/cronjob"
	news "itmrchow/tw-media-analytics-service/domain/news/delivery"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	"itmrchow/tw-media-analytics-service/domain/news/service"
	"itmrchow/tw-media-analytics-service/domain/queue"
	spider "itmrchow/tw-media-analytics-service/domain/spider/delivery"
	"itmrchow/tw-media-analytics-service/infra"
)

func main() {
	// 系統信號處理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// config
	infra.InitConfig()

	// logger
	logger := infra.InitLogger()
	infra.SetInfraLogger(logger)

	// context
	ctx, cancel := context.WithCancel(context.Background())

	// Set up OpenTelemetry
	otelShutdown, err := infra.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	tracer := otel.Tracer("media-analytics-server")
	ctx, span := tracer.Start(ctx, "server init")
	// meter := otel.Meter(name)

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
	newsRepo := repository.NewNewsRepositoryImpl(logger, db)
	authorRepo := repository.NewAuthorRepositoryImpl(logger, db)
	analysisRepo := repository.NewAnalysisRepositoryImpl(logger, db)

	// service
	newsService := service.NewNewsServiceImpl(logger, newsRepo, authorRepo, analysisRepo, q, db, model)

	// handler
	// - Spider handler
	spiderEventHandlerMap := map[uint]*spider.SpiderEventHandler{
		1: spider.NewCtiNewsNewsSpiderEventHandler(logger, q), // 中天
		2: spider.NewSetnNewsSpiderEventHandler(logger, q),    // 三立
	}
	spiderEventHandler := spider.NewBaseEventHandler(logger, spiderEventHandlerMap)

	// - news handler
	newsHandler := news.NewNewsEventHandler(logger, q, db, newsService)

	// consumer
	go func() {
		if err := initConsumer(ctx, q, spiderEventHandler, newsHandler); err != nil {
			log.Err(err).Msg("failed to init consumer")
			cancel()
		}
	}()

	logger.Info().Msg("server started")
	span.End()

	defer func() {
		q.CloseClient()
		model.CloseClient()

		logger.Info().Msg("Client closed")

	}()

	select {
	case sig := <-sigChan:
		logger.Info().Msgf("收到系統信號: %v, 開始關閉服務", sig)
		cancel()
	case <-ctx.Done():
		logger.Info().Msg("服務開始關閉")
	}
}

func initConsumer(ctx context.Context, q queue.Queue,
	spiderEventHandler *spider.BaseEventHandler,
	newsHandler *news.NewsEventHandler,
) (err error) {
	// create error group
	var group errgroup.Group

	// set subscription

	// - ArticleListScraping
	for mediaID, h := range spiderEventHandler.SpiderMap {
		group.Go(func() error {
			subID := fmt.Sprintf(
				"%s_%s_%v_sub",
				string(queue.TopicArticleListScraping),
				viper.GetString("ENV"),
				mediaID,
			)
			return q.Consume(ctx, queue.TopicArticleListScraping, subID, h.ArticleListScrapingHandle)
		})
	}

	// - NewsCheck
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicNewsCheck, "", newsHandler.CheckNewsExistHandle)
	})

	// - ArticleContentScraping
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicArticleContentScraping, "", spiderEventHandler.ArticleContentScrapingHandle)
	})

	// - NewsSave
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicNewsSave, "", newsHandler.SaveNewsHandle)
	})

	// - GetAnalysis
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicGetAnalysis, "", newsHandler.GetAnalysisHandle)
	})

	// - SaveAnalysis

	return group.Wait()
}
