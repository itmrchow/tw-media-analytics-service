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
	msg := utils.GetNewsEvent{}

	if err := c.queue.Publish(ctx, queue.TopicArticleScraping, msg); err != nil {
		c.logger.Error().Err(err).Msg("ArticleScrapingJob Publish Error")
	}

	c.logger.Info().Msg("ArticleScrapingJob End")
}
