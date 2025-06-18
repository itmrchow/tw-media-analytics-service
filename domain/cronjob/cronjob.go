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

// ArticleScrapingJob 觸發爬取文章 pub.
func (c *CronJob) ArticleScrapingJob() {
	// create new context
	ctx := context.Background()

	// Tracer
	ctx, span := c.tracer.Start(ctx, "domain/cronjob/cronjob/ArticleScrapingJob:Article Scraping Job")
	c.logger.Info().Ctx(ctx).Msg("ArticleScrapingJob: start")
	defer func() {
		c.logger.Info().Ctx(ctx).Msg("ArticleScrapingJob: end")
		span.End()
	}()

	// publish
	payload, err := json.Marshal(utils.EventArticleListScraping{})
	if err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msg("ArticleScrapingJob Marshal Error")
		return
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	if err = c.publisher.Publish(string(queue.TopicArticleListScraping), msg); err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msg("ArticleScrapingJob Publish Error")
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

// InitCron 初始化 cron job.
func (c *CronJob) InitCron(ctx context.Context) {
	// Tracer
	ctx, span := c.tracer.Start(ctx, "infra/cronjob/InitCron: Init Cron")
	c.logger.Info().Ctx(ctx).Msg("InitCron: start")
	defer func() {
		span.End()
		c.logger.Info().Ctx(ctx).Msg("InitCron: end")
	}()

	cr := cron.New()
	// ArticleScrapingJob
	_, err := cr.AddFunc("0 * * * *", c.ArticleScrapingJob)
	if err != nil {
		c.logger.Fatal().Err(err).Ctx(ctx).Str("job", "ArticleScrapingJob").Msg("failed to add cron job")
	}

	// AnalyzeNewsJob
	// _, err = c.AddFunc("*/1 * * * *", jobs.AnalyzeNewsJob)
	// if err != nil {
	// 	logger.Fatal().Err(err).Ctx(ctx).Str("job", "AnalyzeNewsJob").Msg("failed to add cron job")
	// }

	cr.Start()
}
