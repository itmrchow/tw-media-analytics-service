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
	tracer := otel.Tracer("domain/cronjob/cronjob")

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

	// tracer
	ctx, span := c.tracer.Start(ctx, "ArticleScrapingJob")
	c.logger.Info().Ctx(ctx).Msg("ArticleScrapingJob: start")
	defer func() {
		span.End()
		c.logger.Info().Ctx(ctx).Msg("ArticleScrapingJob end")
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

	// tracer
	ctx, span := c.tracer.Start(ctx, "AnalyzeNewsJob")
	c.logger.Info().Ctx(ctx).Msg("AnalyzeNewsJob: start")
	defer func() {
		span.End()
		c.logger.Info().Ctx(ctx).Msg("AnalyzeNewsJob end")
	}()

	// publish
	msg := utils.EventNewsAnalysis{
		AnalysisNum: 2, // 設定一次分析2筆 // TODO: 從 config 讀取
	}

	if err := c.queue.Publish(ctx, queue.TopicGetAnalysis, msg); err != nil {
		c.logger.Error().Err(err).Msg("AnalyzeNewsJob Publish Error")
	}
}
