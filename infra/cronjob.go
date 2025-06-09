package infra

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/cronjob"
)

// InitCron 初始化 cron job.
func InitCron(ctx context.Context, logger *zerolog.Logger, jobs *cronjob.CronJob) {
	// Trace
	ctx, span := tracer.Start(ctx, "infra/cronjob:InitCron")
	logger.Info().Ctx(ctx).Msg("InitCron: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitCron: end")
	}()

	c := cron.New()
	// ArticleScrapingJob
	_, err := c.AddFunc("0 * * * *", jobs.ArticleScrapingJob)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Str("job", "ArticleScrapingJob").Msg("failed to add cron job")
	}

	// AnalyzeNewsJob
	_, err = c.AddFunc("*/1 * * * *", jobs.AnalyzeNewsJob)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Str("job", "AnalyzeNewsJob").Msg("failed to add cron job")
	}

	c.Start()
}
