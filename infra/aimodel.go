package infra

import (
	"context"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/ai"
)

func InitAIModel(ctx context.Context, initLogger *zerolog.Logger) ai.AiModel {
	// Trace
	tracer := getInfraTracer()
	ctx, span := tracer.Start(ctx, "InitAIModel")
	logger.Info().Ctx(ctx).Msg("InitAIModel: start")
	defer func() {
		span.End()
		logger.Info().Ctx(ctx).Msg("InitAIModel: end")
	}()

	// New Gemini
	model, err := ai.NewGemini(ctx, initLogger)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("InitAIModel: failed to create Gemini model")
	}

	return model
}
