package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"

	"itmrchow/tw-media-analytics-service/domain/spider/entity"
)

var _ Spider = &SetnSpider{}

// 三立
// 從https://www.setn.com/sitemapGoogleNews.xml 抓列表
// 從列表中抓取新聞

type SetnSpider struct {
	tracer          trace.Tracer
	logger          *zerolog.Logger
	newsPageURL     string
	newsListPageURL string
	goquerySelector string
	mediaID         uint
}

func NewSetnSpider(logger *zerolog.Logger, tracer trace.Tracer) *SetnSpider {
	var spider = &SetnSpider{
		tracer:          tracer,
		logger:          logger,
		newsPageURL:     "https://www.setn.com/News.aspx?NewsID=%d",
		newsListPageURL: "https://www.setn.com/sitemapGoogleNews.xml",
		goquerySelector: "script[type='application/ld+json']",
		mediaID:         2,
	}

	return spider
}

func (s *SetnSpider) GetNews(ctx context.Context, newsID string) (*entity.News, error) {
	// Trace
	ctx, span := s.tracer.Start(
		ctx,
		"domain/spider/usecase/spider_setn/GetNews: Get News",
	)
	defer func() {
		span.End()
		s.logger.Info().Ctx(ctx).Uint("media_id", s.mediaID).Msg("GetNews: end")
	}()

	s.logger.Info().Ctx(ctx).Uint("media_id", s.mediaID).Msg("GetNews: start")

	// 建立新的收集器
	c := colly.NewCollector()

	// 設定請求頭
	c.OnRequest(func(r *colly.Request) {
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
	c.OnResponse(func(r *colly.Response) {
		responseSize = len(r.Body)
	})

	// 處理錯誤
	c.OnError(func(r *colly.Response, err error) {
		s.logger.Error().Err(err).Ctx(ctx).Msgf("Error: %v", err)
	})

	// 處理 HTML - 獲取新聞內容
	c.OnHTML("div#ckuse div#Content1", func(e *colly.HTMLElement) {
		// 獲取新聞內容
		newsData.NewsContext = strings.TrimSpace(e.Text)
	})

	// 處理 HTML
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		// 先解析 type
		var typeCheck struct {
			Type string `json:"@type"`
		}

		err := json.Unmarshal([]byte(e.Text), &typeCheck)
		if err != nil {
			s.logger.Error().Err(err).Ctx(ctx).Msgf("Error parsing JSON: %v", err)
			return
		}

		if typeCheck.Type != "NewsArticle" {
			return
		}

		// 解析 JSON
		err = json.Unmarshal([]byte(e.Text), &newsData)
		if err != nil {
			s.logger.Error().Err(err).Ctx(ctx).Msgf("Error parsing JSON: %v", err)
			return
		}
	})

	// 開始抓取
	url := fmt.Sprintf("https://www.setn.com/News.aspx?NewsID=%s", newsID)
	err := c.Visit(url)
	if err != nil {
		s.logger.Error().Err(err).Ctx(ctx).Msgf("Error visiting URL: %v , URL: %s", err, url)
		return nil, err
	}

	// 計算執行時間
	elapsedTime := time.Since(startTime)

	newsData.ElapsedTime = elapsedTime
	newsData.ResponseSize = responseSize

	s.logger.Info().Ctx(ctx).
		Str("id", newsData.NewsID).
		Str("title", newsData.Headline[:min(10, len(newsData.Headline))]).
		Dur("elapsed_time", elapsedTime).
		Int("response_size", responseSize).
		Msg("News scraping completed , send news save event")

	return &newsData, nil
}

func (s *SetnSpider) GetNewsList(ctx context.Context, newsIDList []string) ([]*entity.News, error) {
	// Trace
	ctx, span := s.tracer.Start(ctx, "domain/spider/usecase/spider_setn/GetNewsList: Get News List")
	defer func() {
		s.logger.Info().Ctx(ctx).Uint("media_id", s.mediaID).Msg("GetNewsList: end")
		span.End()
	}()

	s.logger.Info().Ctx(ctx).Uint("media_id", s.mediaID).Msg("GetNewsList: start")

	newsDataList := make([]*entity.News, 0)

	for _, newsID := range newsIDList {
		newsData, err := s.GetNews(ctx, newsID)
		if err != nil {
			return nil, err
		}
		newsDataList = append(newsDataList, newsData)
	}

	return newsDataList, nil
}

func (s *SetnSpider) GetNewsIdList(ctx context.Context) ([]string, error) {
	// Trace
	ctx, span := s.tracer.Start(ctx, "domain/spider/usecase/spider_setn/GetNewsIdList: Get News ID List")
	defer func() {
		span.End()
		s.logger.Info().Ctx(ctx).Msg("GetNewsIdList: end")
		s.logger.Info().Ctx(ctx).Uint("media_id", s.mediaID).Msg("GetNewsIdList: end")
	}()
	// 建立新的收集器
	c := colly.NewCollector()

	// 儲存新聞ID列表
	var newsIDs []string

	// 設定請求頭
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set(
			"User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		)
	})

	// 處理錯誤
	c.OnError(func(r *colly.Response, err error) {
		s.logger.Error().Err(err).Ctx(ctx).Msgf("Error fetching sitemap: %v", err)
	})

	// 處理 XML
	c.OnXML("//url/loc", func(e *colly.XMLElement) {
		url := e.Text
		// 只處理新聞頁面的 URL
		if strings.Contains(url, "News.aspx") {
			// 從 URL 中提取 NewsID
			parts := strings.Split(url, "NewsID=")
			if len(parts) == 2 {
				newsID := parts[1]

				newsIDs = append(newsIDs, newsID)
			}
		}
	})

	// 開始抓取
	err := c.Visit("https://www.setn.com/sitemapGoogleNews.xml")
	if err != nil {
		s.logger.Error().Err(err).Ctx(ctx).Msgf("error visiting sitemap: %v", err)
		return nil, err
	}

	s.logger.Info().Ctx(ctx).Msgf("三立找到 %d 篇新聞文章", len(newsIDs))

	return newsIDs, nil
}

func (s *SetnSpider) GetMediaID() uint {
	return s.mediaID
}
