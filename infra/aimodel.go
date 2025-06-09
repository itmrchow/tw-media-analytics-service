package infra

import (
	"context"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/ai"
)

func InitAIModel(ctx context.Context, initLogger *zerolog.Logger) ai.AiModel {
	// Trace
	ctx, span := tracer.Start(ctx, "infra/InitAIModel: Init AI Model")
	logger.Info().Ctx(ctx).Msg("InitAIModel: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("InitAIModel: end")
		span.End()
	}()

	// New Gemini
	model, err := ai.NewGemini(ctx, initLogger)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("InitAIModel: failed to create Gemini model")
	}

	return model
}
