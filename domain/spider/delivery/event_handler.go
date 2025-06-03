package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
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

func NewBaseEventHandler(logger *zerolog.Logger, spiderEventHandlerMap map[uint]*SpiderEventHandler) *BaseEventHandler {
	return &BaseEventHandler{
		tracer:    otel.Tracer("domain/spider/delivery"),
		logger:    logger,
		SpiderMap: spiderEventHandlerMap,
	}
}

// ArticleContentScrapingHandle 爬取文章內容.
func (h *BaseEventHandler) ArticleContentScrapingHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(ctx, "ArticleContentScrapingHandle")
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
	tracer trace.Tracer
	logger *zerolog.Logger
	queue  queue.Queue
	spider spider.Spider // usecase
}

// ctinews spider event handler (中天)
func NewCtiNewsNewsSpiderEventHandler(logger *zerolog.Logger, queue queue.Queue) *SpiderEventHandler {
	tracer := otel.Tracer("domain/spider/delivery")
	spider := spider.NewCtiNewsSpider(logger, queue)

	return &SpiderEventHandler{
		tracer: tracer,
		logger: logger,
		spider: spider,
		queue:  queue,
	}
}

// setn spider event handler (三立)
func NewSetnNewsSpiderEventHandler(logger *zerolog.Logger, queue queue.Queue) *SpiderEventHandler {
	tracer := otel.Tracer("domain/spider/delivery")
	spider := spider.NewSetnSpider(logger)

	return &SpiderEventHandler{
		tracer: tracer,
		logger: logger,
		spider: spider,
		queue:  queue,
	}
}

// ArticleListScrapingHandle 爬取文章列表.
func (h *SpiderEventHandler) ArticleListScrapingHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(ctx, "ArticleListScrapingHandle")
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
	err = h.queue.Publish(ctx, queue.TopicNewsCheck, checkNewsEvent)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to publish create news event")
		return err
	}

	return nil
}

// ArticleContentScrapingHandle 爬取文章內容.
func (h *SpiderEventHandler) ArticleContentScrapingHandle(ctx context.Context, msg []byte) error {
	// Tracer
	ctx, span := h.tracer.Start(ctx, "ArticleContentScrapingHandle")
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

	err = h.queue.Publish(ctx, queue.TopicNewsSave, checkNewsEvent)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to publish create news event")
		return err
	}

	checkNewsEventJSON, err := json.Marshal(checkNewsEvent)
	if err != nil {
		h.logger.Error().Err(err).Ctx(ctx).Msg("failed to marshal checkNewsEvent")
		return err
	}

	h.logger.Info().Ctx(ctx).
		Str("check_news_event", string(checkNewsEventJSON)).
		Msg("check_news_event")

	return nil
}
