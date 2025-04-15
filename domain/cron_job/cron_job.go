package cron_job

import "github.com/rs/zerolog"

type CronJob struct {
	logger *zerolog.Logger
}

func NewCronJob(logger *zerolog.Logger) *CronJob {
	return &CronJob{logger: logger}
}

func (c *CronJob) ArticleScrapingJob() {
	c.logger.Info().Msg("ArticleScrapingJob Start")

	// TODO: send event to kafka
	// TODO: pub sub to kafka
	c.logger.Info().Msg("ArticleScrapingJob End")
}
