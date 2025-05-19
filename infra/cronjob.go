package infra

import (
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/cronjob"
)

func InitCron(logger *zerolog.Logger, jobs *cronjob.CronJob) {

	c := cron.New()
	// ArticleScrapingJob
	_, err := c.AddFunc("0 * * * *", jobs.ArticleScrapingJob)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to add cron job")
	}

	// AnalyzeNewsJob
	_, err = c.AddFunc("*/1 * * * *", jobs.AnalyzeNewsJob)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to add cron job")
	}

	c.Start()
	logger.Info().Msg("cron job started")
}
