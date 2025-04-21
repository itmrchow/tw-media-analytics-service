package cron_job

import (
	"context"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type CronJob struct {
	logger *zerolog.Logger
	queue  queue.Queue
}

func NewCronJob(logger *zerolog.Logger, queue queue.Queue) *CronJob {

	return &CronJob{
		logger: logger,
		queue:  queue,
	}
}

// 觸發爬取文章 pub
func (c *CronJob) ArticleScrapingJob() {
	c.logger.Info().Msg("ArticleScrapingJob Start")

	ctx := context.Background()
	msg := utils.EventArticleListScraping{}

	if err := c.queue.Publish(ctx, queue.TopicArticleListScraping, msg); err != nil {
		c.logger.Error().Err(err).Msg("ArticleScrapingJob Publish Error")
	}

	c.logger.Info().Msg("ArticleScrapingJob End")
}

// 觸發分析文章 pub
func (c *CronJob) AnalyzeNewsJob() {
	c.logger.Info().Msg("AnalyzeNewsJob Start")

	ctx := context.Background()
	msg := utils.EventNewsAnalysis{
		AnalysisNum: 10, // 設定一次分析10筆
	}

	if err := c.queue.Publish(ctx, queue.TopicGetAnalysis, msg); err != nil {
		c.logger.Error().Err(err).Msg("AnalyzeNewsJob Publish Error")
	}

	c.logger.Info().Msg("AnalyzeNewsJob End")
}
