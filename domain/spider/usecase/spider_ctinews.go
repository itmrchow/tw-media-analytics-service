package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/spider/entity"
)

var _ Spider = &CtiNewsSpider{}

type CtiNewsSpider struct {
	tracer          trace.Tracer
	logger          *zerolog.Logger
	newsPageURL     string
	newsListPageURL string
	goquerySelector string
	mediaID         uint
}

func NewCtiNewsSpider(logger *zerolog.Logger, queue queue.Queue) *CtiNewsSpider {
	var spider = &CtiNewsSpider{
		tracer:          otel.Tracer("domain/spider/usecase"),
		logger:          logger,
		newsPageURL:     "https://ctinews.com/news/items/%s",
		newsListPageURL: "https://ctinews.com/rss/sitemap-news.xml",
		goquerySelector: "script[type='application/ld+json']",
		mediaID:         1,
	}

	return spider
}

func (c *CtiNewsSpider) GetNews(ctx context.Context, newsID string) (*entity.News, error) {
	// Trace
	ctx, span := c.tracer.Start(ctx, "GetNews")
	defer func() {
		span.End()
		c.logger.Info().Ctx(ctx).Uint("media_id", c.mediaID).Msg("GetNews: end")
	}()

	c.logger.Info().Ctx(ctx).Uint("media_id", c.mediaID).Msg("GetNews: start")

	// 建立新的收集器
	collector := colly.NewCollector()

	// 設定請求頭
	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		)
	})

	// 記錄開始時間
	startTime := time.Now()

	// 儲存回應大小
	var responseSize int

	// 儲存新聞資料
	var newsData entity.News
	newsData.NewsID = newsID

	// 處理回應
	collector.OnResponse(func(r *colly.Response) {
		responseSize = len(r.Body)
	})

	// 處理錯誤
	collector.OnError(func(r *colly.Response, err error) {
		c.logger.Error().Err(err).Ctx(ctx).Msg("錯誤")
	})

	// 處理 JSON 資料
	collector.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {

		// 先解析 type
		var NewsArticle struct {
			Type     string `json:"@type"`
			Headline string `json:"headline"`
			Content  string `json:"articleBody,omitempty"`
			Author   struct {
				Type string `json:"@type"`
				Name string `json:"name"`
			} `json:"author"`
			Category      string    `json:"articleSection"`
			URL           string    `json:"url"`
			DatePublished time.Time `json:"datePublished"`
			DateModified  time.Time `json:"dateModified"`
		}

		err := json.Unmarshal([]byte(e.Text), &NewsArticle)
		if err != nil {
			c.logger.Error().Err(err).Ctx(ctx).Msg("解析 JSON 錯誤")
			return
		}

		if NewsArticle.Type != "NewsArticle" {
			return
		}

		// 解析 JSON
		newsData.NewsID = newsID
		newsData.Headline = NewsArticle.Headline
		newsData.Author.Type = NewsArticle.Author.Type
		newsData.Author.Name = NewsArticle.Author.Name
		newsData.DatePublished = NewsArticle.DatePublished
		newsData.DateModified = NewsArticle.DateModified
		newsData.NewsContext = NewsArticle.Content
		newsData.Category = NewsArticle.Category
		newsData.URL = NewsArticle.URL

		newsData.ResponseSize = responseSize
		newsData.ElapsedTime = time.Since(startTime)
	})

	// 開始抓取
	url := fmt.Sprintf(c.newsPageURL, newsID)
	err := collector.Visit(url)
	if err != nil {
		c.logger.Error().Err(err).Ctx(ctx).Msgf("訪問 URL 錯誤: %v, URL: %s", err, url)
		return nil, err
	}

	// 計算執行時間
	elapsedTime := time.Since(startTime)

	newsData.ElapsedTime = elapsedTime
	newsData.ResponseSize = responseSize

	c.logger.Info().Ctx(ctx).
		Str("id", newsData.NewsID).
		Str("title", newsData.Headline[:min(10, len(newsData.Headline))]).
		Dur("elapsed_time", elapsedTime).
		Int("response_size", responseSize).
		Msg("News scraping completed")

	return &newsData, nil
}

func (c *CtiNewsSpider) GetNewsList(ctx context.Context, newsIDList []string) ([]*entity.News, error) {
	// Trace
	ctx, span := c.tracer.Start(ctx, "GetNewsList")
	defer func() {
		span.End()
		c.logger.Info().Ctx(ctx).Msg("GetNewsList: end")
		c.logger.Info().Ctx(ctx).Uint("media_id", c.mediaID).Msg("GetNewsList: end")
	}()

	c.logger.Info().Ctx(ctx).Msg("GetNewsList: start")
	newsDataList := make([]*entity.News, 0)

	for _, newsID := range newsIDList {
		newsData, err := c.GetNews(ctx, newsID)
		if err != nil {
			return nil, err
		}
		newsDataList = append(newsDataList, newsData)
	}

	return newsDataList, nil
}

func (c *CtiNewsSpider) GetNewsIdList(ctx context.Context) ([]string, error) {
	// Trace
	ctx, span := c.tracer.Start(ctx, "GetNewsIdList")
	defer func() {
		span.End()
		c.logger.Info().Ctx(ctx).Msg("GetNewsIdList: end")
		c.logger.Info().Ctx(ctx).Uint("media_id", c.mediaID).Msg("GetNewsIdList: end")
	}()

	// 建立新的收集器
	collector := colly.NewCollector()

	// 儲存新聞ID列表
	var newsIDs []string

	// 設定請求頭
	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		)
	})

	// 處理錯誤
	collector.OnError(func(r *colly.Response, err error) {
		log.Printf("取得網站地圖錯誤: %v", err)
	})

	// 處理 XML
	collector.OnXML("//url/loc", func(e *colly.XMLElement) {
		url := e.Text
		// 只處理新聞頁面的 URL
		if strings.Contains(url, "ctinews.com/news/items/") {
			// 從 URL 中提取 NewsID
			parts := strings.Split(url, "/items/")
			if len(parts) == 2 {
				newsID := parts[1]
				newsIDs = append(newsIDs, newsID)
			}
		}
	})

	// 開始抓取
	err := collector.Visit(c.newsListPageURL)
	if err != nil {
		return nil, fmt.Errorf("訪問網站地圖錯誤: %v", err)
	}

	c.logger.Info().Msgf("中天找到 %d 篇新聞文章", len(newsIDs))

	return newsIDs, nil
}

func (c *CtiNewsSpider) GetMediaID() uint {
	return c.mediaID
}
