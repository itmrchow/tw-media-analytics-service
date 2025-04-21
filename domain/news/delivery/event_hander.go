package delivery

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/service"
	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type NewsEventHandler struct {
	log   *zerolog.Logger
	queue queue.Queue
	db    *gorm.DB

	newsService service.NewsService
}

func NewNewsEventHandler(log *zerolog.Logger, queue queue.Queue, db *gorm.DB, newsService service.NewsService) *NewsEventHandler {
	return &NewsEventHandler{log: log, queue: queue, newsService: newsService, db: db}
}

func (h *NewsEventHandler) CheckNewsExistHandle(ctx context.Context, msg []byte) error {

	// check msg event type
	var checkNewsEvent utils.EventNewsCheck
	if err := json.Unmarshal(msg, &checkNewsEvent); err != nil {
		h.log.Error().Err(err).Msg("failed to unmarshal message to CheckNewsEvent")
		return err
	}

	// check news id exist in db
	if err := h.newsService.CheckNewsExist(ctx, checkNewsEvent); err != nil {
		h.log.Error().Err(err).Msg("failed to check news exist")
		return err
	}

	return nil
}

func (h *NewsEventHandler) SaveNewsHandle(ctx context.Context, msg []byte) error {

	// check msg event type
	var saveNewsEvent utils.EventNewsSave
	if err := json.Unmarshal(msg, &saveNewsEvent); err != nil {
		h.log.Error().Err(err).Msg("failed to unmarshal message to SaveNewsEvent")
		return err
	}

	if err := h.newsService.SaveNews(ctx, saveNewsEvent); err != nil {
		h.log.Error().Err(err).Msg("failed to save news")
		return err
	}

	return nil
}

func (h *NewsEventHandler) GetAnalysisHandle(ctx context.Context, msg []byte) error {

	// check msg event type
	var checkNewsEvent utils.EventNewsAnalysis
	if err := json.Unmarshal(msg, &checkNewsEvent); err != nil {
		h.log.Error().Err(err).Msg("failed to unmarshal message to GetAnalysisEvent")
		return err
	}

	h.log.Info().Msg("GetAnalysisHandle")

	if err := h.newsService.AnalysisNews(ctx, checkNewsEvent); err != nil {
		h.log.Error().Err(err).Msg("failed to analysis news")
		return err
	}

	return nil
}
