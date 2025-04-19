package delivery

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
	spider "itmrchow/tw-media-analytics-service/domain/spider/usecase"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type SpiderEventHandler struct {
	log    *zerolog.Logger
	spider spider.Spider // usecase
	queue  queue.Queue
}

// ctinews spider event handler (中天)
func NewCtiNewsNewsSpiderEventHandler(log *zerolog.Logger, queue queue.Queue) *SpiderEventHandler {
	spider := spider.NewCtiNewsSpider(log, queue)

	return &SpiderEventHandler{
		log:    log,
		spider: spider,
		queue:  queue,
	}
}

// setn spider event handler (三立)
func NewSetnNewsSpiderEventHandler(log *zerolog.Logger, queue queue.Queue) *SpiderEventHandler {
	spider := spider.NewSetnSpider(log)

	return &SpiderEventHandler{
		log:    log,
		spider: spider,
		queue:  queue,
	}
}

func (h *SpiderEventHandler) ArticleListScrapingHandle(ctx context.Context, msg []byte) error {

	h.log.Info().Msgf("ArticleListScrapingHandle, mediaID: %v, msg: %s", h.spider.GetMediaID(), string(msg))

	var event utils.EventArticleListScraping
	if err := json.Unmarshal(msg, &event); err != nil {
		h.log.Error().Err(err).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	// get news id list
	newsIDList, err := h.spider.GetNewsIdList()
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get news id list")
		return err
	}

	h.log.Info().Msgf("Found %d news", len(newsIDList))

	// publish check news event
	checkNewsEvent := utils.EventNewsCheck{
		MediaID:    h.spider.GetMediaID(),
		NewsIDList: newsIDList,
	}
	err = h.queue.Publish(ctx, queue.TopicNewsCheck, checkNewsEvent)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to publish create news event")
		return err
	}

	return nil
}

func (h *SpiderEventHandler) ArticleContentScrapingHandle(ctx context.Context, msg []byte) error {
	h.log.Info().Msgf("ArticleContentScrapingHandle ctinews: %s", string(msg))
	return nil
}
