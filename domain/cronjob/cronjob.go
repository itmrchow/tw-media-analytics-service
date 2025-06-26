package cronjob

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type CronJob struct {
	tracer    trace.Tracer
	logger    *zerolog.Logger
	publisher message.Publisher
}

func NewCronJob(logger *zerolog.Logger, tracer trace.Tracer, publisher message.Publisher) *CronJob {
	return &CronJob{
		tracer:    tracer,
		logger:    logger,
		publisher: publisher,
	}
}

// NewsListScrapingJob 觸發爬取新聞列表 pub.
func (c *CronJob) NewsListScrapingJob() {
	// create new context
	ctx := context.Background()

	// Tracer
	ctx, span := c.tracer.Start(ctx, "domain/cronjob/cronjob/NewsListScrapingJob:News List Scraping Job")
	c.logger.Info().Ctx(ctx).Msg("NewsListScrapingJob: start")
	defer func() {
		c.logger.Info().Ctx(ctx).Msg("NewsListScrapingJob: end")
		span.End()
	}()

	// prepare event payload
	payload, err := json.Marshal(utils.EventNewsListScraping{})
	if err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msg("NewsListScrapingJob Marshal Error")
		return
	}

	// publish
	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err = c.publisher.Publish(
		string(queue.TopicNewsListScraping),
		msg,
	); err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msg("NewsListScrapingJob Publish Error")
	}
}

// AnalyzeNewsJob 觸發分析文章 pub.
func (c *CronJob) AnalyzeNewsJob() {
	// create new context
	ctx := context.Background()

	// Tracer
	ctx, span := c.tracer.Start(ctx, "domain/cronjob/cronjob/AnalyzeNewsJob:Analyze News Job")
	c.logger.Info().Ctx(ctx).Msg("AnalyzeNewsJob: start")
	defer func() {
		c.logger.Info().Ctx(ctx).Msg("AnalyzeNewsJob end")
		span.End()
	}()

	// publish
	payload, err := json.Marshal(utils.EventNewsAnalysis{
		AnalysisNum: 2, // 設定一次分析2筆 // TODO: 從 config 讀取
	})
	if err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msg("AnalyzeNewsJob Marshal Error")
		return
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err = c.publisher.Publish(string(queue.TopicGetAnalysis), msg); err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msg("AnalyzeNewsJob Publish Error")
	}
}

// InitCronJob 初始化 cron job.
func InitCronJob(ctx context.Context, logger *zerolog.Logger, tracer trace.Tracer, cronJob *CronJob) {
	// Tracer
	ctx, span := tracer.Start(ctx, "domain/cronjob/cronjob/InitCron: Init Cron")
	logger.Info().Ctx(ctx).Msg("InitCronJob: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitCronJob: end")
	}()

	cr := cron.New()
	// ArticleScrapingJob
	_, err := cr.AddFunc("0 * * * *", cronJob.NewsListScrapingJob)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Str("job", "ArticleScrapingJob").Msg("failed to add cron job")
	}

	// AnalyzeNewsJob
	_, err = cr.AddFunc("*/1 * * * *", cronJob.AnalyzeNewsJob)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Str("job", "AnalyzeNewsJob").Msg("failed to add cron job")
	}

	cr.Start()
}
