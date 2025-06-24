package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/queue"
	spider "itmrchow/tw-media-analytics-service/domain/spider/usecase"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

type BaseEventHandler struct {
	tracer    trace.Tracer
	logger    *zerolog.Logger
	SpiderMap map[uint]*SpiderEventHandler
}

func NewBaseEventHandler(
	logger *zerolog.Logger,
	tracer trace.Tracer,
	spiderHandlers []*SpiderEventHandler,
) *BaseEventHandler {

	m := make(map[uint]*SpiderEventHandler)
	for _, v := range spiderHandlers {
		m[v.spider.GetMediaID()] = v
	}

	return &BaseEventHandler{
		tracer:    tracer,
		logger:    logger,
		SpiderMap: m,
	}
}

// ArticleContentScrapingHandle 爬取文章內容.
func (h *BaseEventHandler) ArticleContentScrapingHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(
		ctx,
		"domain/spider/delivery/event_handler/ArticleContentScrapingHandle: Article Content Scraping Handle",
	)
	h.logger.Info().Ctx(ctx).Msg("ArticleContentScrapingHandle: start")
	defer func() {
		span.End()
		h.logger.Info().Ctx(ctx).Msg("ArticleContentScrapingHandle end")
	}()

	var event utils.EventArticleContentScraping
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	spiderHandler, ok := h.SpiderMap[event.MediaID]
	if !ok {
		h.logger.Error().Ctx(ctx).Msgf("spider handler not found, mediaID: %v", event.MediaID)
		return fmt.Errorf("spider handler not found, mediaID: %v", event.MediaID)
	}

	return spiderHandler.ArticleContentScrapingHandle(ctx, msg)
}

type SpiderEventHandler struct {
	tracer    trace.Tracer
	logger    *zerolog.Logger
	publisher message.Publisher
	spider    spider.Spider // usecase
}

// ctinews spider event handler (中天)
func NewCtiNewsNewsSpiderEventHandler(
	logger *zerolog.Logger,
	tracer trace.Tracer,
	publisher message.Publisher,
	spider spider.Spider,
) *SpiderEventHandler {

	return &SpiderEventHandler{
		tracer:    tracer,
		logger:    logger,
		publisher: publisher,
		spider:    spider,
	}
}

// setn spider event handler (三立)
func NewSetnNewsSpiderEventHandler(
	logger *zerolog.Logger,
	tracer trace.Tracer,
	publisher message.Publisher,
	spider spider.Spider,
) *SpiderEventHandler {

	return &SpiderEventHandler{
		tracer:    tracer,
		logger:    logger,
		publisher: publisher,
		spider:    spider,
	}
}

// ArticleListScrapingHandle 爬取文章列表.
func (h *SpiderEventHandler) ArticleListScrapingHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(
		ctx,
		"domain/spider/delivery/event_handler/ArticleListScrapingHandle: Article List Scraping Handle",
	)
	h.logger.Info().Ctx(ctx).
		Uint("media_id", h.spider.GetMediaID()).
		Msg("ArticleListScrapingHandle: start")
	defer func() {
		span.End()
		h.logger.Info().Ctx(ctx).Msg("ArticleListScrapingHandle end")
	}()

	h.logger.Info().Ctx(ctx).
		Uint("media_id", h.spider.GetMediaID()).
		Str("msg", string(msg))

	var event utils.EventArticleListScraping
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	// get news id list
	newsIDList, err := h.spider.GetNewsIdList(ctx)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to get news id list")
		return err
	}

	h.logger.Info().Ctx(ctx).
		Uint("media_id", h.spider.GetMediaID()).
		Int("news_id_list_len", len(newsIDList)).
		Msg("Found news")

	// publish check news event
	checkNewsEvent := utils.EventNewsCheck{
		MediaID:    h.spider.GetMediaID(),
		NewsIDList: newsIDList,
	}

	jsonData, err := json.Marshal(checkNewsEvent)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to marshal checkNewsEvent")
		return err
	}

	checkNewsEventMsg := message.NewMessage(watermill.NewUUID(), jsonData)
	checkNewsEventMsg.SetContext(ctx)

	err = h.publisher.Publish(string(queue.TopicNewsCheck), checkNewsEventMsg)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to publish create news event")
		return err
	}

	return nil
}

// ArticleContentScrapingHandle 爬取文章內容.
func (h *SpiderEventHandler) ArticleContentScrapingHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(
		ctx,
		"domain/spider/delivery/event_handler/ArticleContentScrapingHandle: Article Content Scraping Handle",
	)
	h.logger.Info().Ctx(ctx).
		Uint("media_id", h.spider.GetMediaID()).
		Msg("ArticleContentScrapingHandle: start")
	defer func() {
		span.End()
		h.logger.Info().Ctx(ctx).Msg("ArticleContentScrapingHandle end")
	}()

	var event utils.EventArticleContentScraping
	if err := json.Unmarshal(msg, &event); err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	news, err := h.spider.GetNews(ctx, event.NewsID)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to get news")
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

	jsonData, err := json.Marshal(checkNewsEvent)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to marshal checkNewsEvent")
		return err
	}

	checkNewsEventMsg := message.NewMessage(watermill.NewUUID(), jsonData)
	checkNewsEventMsg.SetContext(ctx)

	err = h.publisher.Publish(string(queue.TopicNewsSave), checkNewsEventMsg)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to publish create news event")
		return err
	}

	h.logger.Info().Ctx(ctx).
		Str("check_news_event", string(jsonData)).
		Msg("check_news_event")

	return nil
}
