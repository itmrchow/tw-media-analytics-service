package cronjob

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type CronJob struct {
	tracer trace.Tracer
	logger *zerolog.Logger
	queue  queue.Queue
}

func NewCronJob(logger *zerolog.Logger, queue queue.Queue) *CronJob {
	// Tracer
	tracer := otel.Tracer("/domain/cronjob/cronjob:NewCronJob")

	return &CronJob{
		tracer: tracer,
		logger: logger,
		queue:  queue,
	}
}

// ArticleScrapingJob 觸發爬取文章 pub.
func (c *CronJob) ArticleScrapingJob() {
	// create new context
	ctx := context.Background()

	// Tracer
	ctx, span := c.tracer.Start(ctx, "domain/cronjob/cronjob/ArticleScrapingJob:Article Scraping Job")
	c.logger.Debug().Msgf("ArticleScrapingJob: traceID: %s", span.SpanContext().TraceID())
	c.logger.Info().Ctx(ctx).Msg("ArticleScrapingJob: start")
	defer func() {
		c.logger.Info().Ctx(ctx).Msg("ArticleScrapingJob end")
		span.End()
	}()

	// publish
	msg := utils.EventArticleListScraping{}
	if err := c.queue.Publish(ctx, queue.TopicArticleListScraping, msg); err != nil {
		c.logger.Error().Err(err).Msg("ArticleScrapingJob Publish Error")
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
	msg := utils.EventNewsAnalysis{
		AnalysisNum: 2, // 設定一次分析2筆 // TODO: 從 config 讀取
	}

	if err := c.queue.Publish(ctx, queue.TopicGetAnalysis, msg); err != nil {
		c.logger.Error().Err(err).Msg("AnalyzeNewsJob Publish Error")
	}
}
