package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"itmrchow/tw-media-analytics-service/domain/ai_model"
	"itmrchow/tw-media-analytics-service/domain/cron_job"
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
	logger := initLogger()

	ctx, cancel := context.WithCancel(context.Background())

	// ai model
	model := ai_model.NewGemini(logger, ctx)

	// db
	db := infra.InitMysqlDb()

	// queue
	q := initQueue(ctx, logger)

	// cron
	initCron(logger, q)

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

func initLogger() *zerolog.Logger {
	var writer io.Writer

	if viper.GetString("ENV") == "dev" {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("message=%s", i)
			},
			// 設定為 true 會讓 JSON 格式化輸出
			NoColor: false, // 設定為 true 會關閉顏色
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
		}
	} else {
		writer = os.Stdout
	}

	// TODO: setting log level
	logger := zerolog.New(writer).Level(zerolog.DebugLevel)
	logger = logger.With().
		Str("service", "tw-media-analytics-service").
		Time("time", time.Now()).
		Caller().
		Logger()
	return &logger
}

func initCron(logger *zerolog.Logger, queue queue.Queue) {

	jobs := cron_job.NewCronJob(logger, queue)

	c := cron.New()
	// ArticleScrapingJob
	_, err := c.AddFunc("0 * * * *", jobs.ArticleScrapingJob)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add cron job")
	}

	// AnalyzeNewsJob
	_, err = c.AddFunc("*/1 * * * *", jobs.AnalyzeNewsJob)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to add cron job")
	}

	c.Start()
	logger.Info().Msg("cron job started")
}

func initQueue(ctx context.Context, logger *zerolog.Logger) queue.Queue {

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
			subID := fmt.Sprintf("%s_%s_%v_sub", string(queue.TopicArticleListScraping), viper.GetString("ENV"), mediaID)
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
