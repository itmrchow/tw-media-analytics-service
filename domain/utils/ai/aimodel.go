package ai

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/ai"
)

func NewGemini(ctx context.Context, logger *zerolog.Logger, tracer trace.Tracer) ai.AiModel {
	// Trace
	ctx, span := tracer.Start(ctx, "utils/ai/NewGemini: New Gemini")
	logger.Info().Ctx(ctx).Msg("NewGemini: start")
	defer func() {
		logger.Info().Ctx(ctx).Msg("NewGemini: end")
		span.End()
	}()

	// New Gemini
	model, err := ai.NewGemini(ctx, logger)
	if err != nil {
		logger.Fatal().Err(err).Ctx(ctx).Msg("InitAIModel: failed to create Gemini model")
	}

	return model
}
