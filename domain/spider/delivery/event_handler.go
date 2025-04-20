package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"

	"itmrchow/tw-media-analytics-service/domain/queue"
	spider "itmrchow/tw-media-analytics-service/domain/spider/usecase"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type BaseEventHandler struct {
	log       *zerolog.Logger
	SpiderMap map[uint]*SpiderEventHandler
}

func NewBaseEventHandler(log *zerolog.Logger, spiderEventHandlerMap map[uint]*SpiderEventHandler) *BaseEventHandler {
	return &BaseEventHandler{
		log:       log,
		SpiderMap: spiderEventHandlerMap,
	}
}

func (h *BaseEventHandler) ArticleContentScrapingHandle(ctx context.Context, msg []byte) error {

	var event utils.EventArticleContentScraping
	if err := json.Unmarshal(msg, &event); err != nil {
		h.log.Error().Err(err).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	spiderHandler, ok := h.SpiderMap[event.MediaID]
	if !ok {
		h.log.Error().Msgf("spider handler not found, mediaID: %v", event.MediaID)
		return fmt.Errorf("spider handler not found, mediaID: %v", event.MediaID)
	}

	return spiderHandler.ArticleContentScrapingHandle(ctx, msg)
}

type SpiderEventHandler struct {
	log    *zerolog.Logger
	queue  queue.Queue
	spider spider.Spider // usecase
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
	var event utils.EventArticleContentScraping
	if err := json.Unmarshal(msg, &event); err != nil {
		h.log.Error().Err(err).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	news, err := h.spider.GetNews(event.NewsID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get news")
		return err
	}

	// publish check news event
	checkNewsEvent := utils.EventNewsSave{
		MediaID:     h.spider.GetMediaID(),
		NewsID:      event.NewsID,
		Title:       news.Headline,
		Content:     news.NewsContext,
		URL:         news.URL,
		AuthorName:  news.Author.Name,
		PublishedAt: news.DatePublished,
		Category:    news.Category,
	}

	err = h.queue.Publish(ctx, queue.TopicNewsSave, checkNewsEvent)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to publish create news event")
		return err
	}

	checkNewsEventJSON, err := json.Marshal(checkNewsEvent)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to marshal checkNewsEvent")
		return err
	}
	h.log.Info().Msgf("checkNewsEvent: %s", string(checkNewsEventJSON))
	h.log.Info().Msgf("ArticleContentScrapingHandle ctinews: %s", string(msg))
	return nil
}
