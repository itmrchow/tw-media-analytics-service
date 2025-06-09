package delivery

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/service"
	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type NewsEventHandler struct {
	tracer trace.Tracer
	logger *zerolog.Logger
	queue  queue.Queue
	db     *gorm.DB

	newsService service.NewsService
}

func NewNewsEventHandler(
	logger *zerolog.Logger,
	queue queue.Queue,
	db *gorm.DB,
	newsService service.NewsService,
) *NewsEventHandler {
	return &NewsEventHandler{
		tracer:      otel.Tracer("domain/news/delivery"),
		logger:      logger,
		queue:       queue,
		newsService: newsService,
		db:          db,
	}
}

func (h *NewsEventHandler) CheckNewsExistHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(ctx, "domain/news/delivery/event_hander/CheckNewsExistHandle: Check News Exist Handle")
	h.logger.Info().Ctx(ctx).Msg("CheckNewsExistHandle: start")
	defer func() {
		span.End()
		h.logger.Info().Ctx(ctx).Msg("CheckNewsExistHandle end")
	}()

	// check msg event type
	var checkNewsEvent utils.EventNewsCheck
	if err := json.Unmarshal(msg, &checkNewsEvent); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to unmarshal message to CheckNewsEvent")
		return err
	}

	// check news id exist in db
	if err := h.newsService.CheckNewsExist(ctx, checkNewsEvent); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to check news exist")
		return err
	}

	return nil
}

func (h *NewsEventHandler) SaveNewsHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(
		ctx,
		"domain/news/delivery/event_hander/SaveNewsHandle: Save News Handle",
	)
	h.logger.Info().Ctx(ctx).Msg("SaveNewsHandle: start")
	defer func() {
		span.End()
		h.logger.Info().Ctx(ctx).Msg("SaveNewsHandle end")
	}()

	// check msg event type
	var saveNewsEvent utils.EventNewsSave
	if err := json.Unmarshal(msg, &saveNewsEvent); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to unmarshal message to SaveNewsEvent")
		return err
	}

	if err := h.newsService.SaveNews(ctx, saveNewsEvent); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to save news")
		return err
	}

	return nil
}

// GetAnalysisHandle 取得分析結果.
func (h *NewsEventHandler) GetAnalysisHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(
		ctx,
		"domain/news/delivery/event_hander/GetAnalysisHandle: Get Analysis Handle",
	)
	h.logger.Info().Ctx(ctx).Msg("GetAnalysisHandle: start")
	defer func() {
		span.End()
		h.logger.Info().Ctx(ctx).Msg("GetAnalysisHandle end")
	}()

	// check msg event type
	var checkNewsEvent utils.EventNewsAnalysis
	if err := json.Unmarshal(msg, &checkNewsEvent); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to unmarshal message to GetAnalysisEvent")
		return err
	}

	if err := h.newsService.AnalysisNews(ctx, checkNewsEvent); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to analysis news")
		return err
	}

	return nil
}
