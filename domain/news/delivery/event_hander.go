package delivery

import (
	"context"
	"encoding/json"
	"strconv"

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
	nonExistingNewsIDs, err := h.newsService.CheckNewsExist(ctx, checkNewsEvent)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to check news exist")
		return err
	}

	h.log.Info().
		Str("mediaID", strconv.Itoa(int(checkNewsEvent.MediaID))).
		Uint("newsIDSize", uint(len(nonExistingNewsIDs))).
		Msg("check news exist event")

	if len(nonExistingNewsIDs) == 0 {
		h.log.Info().Msg("no news id to save")
		return nil
	}

	// publish
	for _, newsID := range nonExistingNewsIDs {

		scrapingContentEvent := utils.EventTopicArticleContentScraping{
			MediaID: checkNewsEvent.MediaID,
			NewsID:  newsID,
		}

		err = h.queue.Publish(ctx, queue.TopicArticleContentScraping, scrapingContentEvent)
		if err != nil {
			h.log.Error().Err(err).Msg("failed to publish news save event")
			return nil // 不影響其他新聞爬取
		}
	}

	h.log.Info().
		Str("mediaID", strconv.Itoa(int(checkNewsEvent.MediaID))).
		Uint("newsIDSize", uint(len(nonExistingNewsIDs))).
		Msg("send news save")

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
