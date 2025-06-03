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
	model := ai.NewGemini(initLogger, ctx)

	return model
}
