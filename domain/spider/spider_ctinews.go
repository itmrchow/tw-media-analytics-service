package spider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
)

var _ Spider = &CtiNewsSpider{}

type CtiNewsSpider struct {
	log             *zerolog.Logger
	newsPageURL     string
	newsListPageURL string
	goquerySelector string
}

func NewCtiNewsSpider(log *zerolog.Logger) *CtiNewsSpider {
	var spider = &CtiNewsSpider{
		log:             log,
		newsPageURL:     "https://ctinews.com/news/items/%s",
		newsListPageURL: "https://ctinews.com/rss/sitemap-news.xml",
		goquerySelector: "script[type='application/ld+json']",
	}

	return spider
}

func (c *CtiNewsSpider) GetNews(newsID string) (*NewsData, error) {
	// 建立新的收集器
	collector := colly.NewCollector()

	// 設定請求頭
	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	})

	// 記錄開始時間
	startTime := time.Now()

	// 儲存回應大小
	var responseSize int

	// 儲存新聞資料
	var newsData NewsData
	newsData.NewsID = newsID

	// 處理回應
	collector.OnResponse(func(r *colly.Response) {
		responseSize = len(r.Body)
	})

	// 處理錯誤
	collector.OnError(func(r *colly.Response, err error) {
		log.Printf("錯誤: %v", err)
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
			DatePublished time.Time `json:"datePublished"`
			DateModified  time.Time `json:"dateModified"`
		}

		err := json.Unmarshal([]byte(e.Text), &NewsArticle)
		if err != nil {
			log.Printf("解析 JSON 錯誤: %v", err)
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

		newsData.ResponseSize = responseSize
		newsData.ElapsedTime = time.Since(startTime)
	})

	// 開始抓取
	url := fmt.Sprintf(c.newsPageURL, newsID)
	err := collector.Visit(url)
	if err != nil {
		log.Printf("訪問 URL 錯誤: %v, URL: %s", err, url)
		return nil, err
	}

	// 計算執行時間
	elapsedTime := time.Since(startTime)

	// 輸出結果
	log.Println("=== 新聞資訊 ===")
	log.Printf("新聞 ID: %s", newsData.NewsID)
	log.Printf("標題: %s", newsData.Headline)
	log.Printf("內容: %s", newsData.NewsContext)
	log.Printf("作者類型: %s", newsData.Author.Type)
	log.Printf("作者名稱: %s", newsData.Author.Name)
	log.Printf("發布時間: %s", newsData.DatePublished.Format("2006-01-02 15:04:05"))
	log.Printf("修改時間: %s", newsData.DateModified.Format("2006-01-02 15:04:05"))

	log.Println("\n=== 執行資訊 ===")

	newsData.ElapsedTime = elapsedTime
	newsData.ResponseSize = responseSize

	log.Printf("花費時間: %v", elapsedTime)
	log.Printf("回應大小: %d bytes", responseSize)

	return &newsData, nil
}

func (c *CtiNewsSpider) GetNewsList(newsIDList []string) ([]*NewsData, error) {
	newsDataList := make([]*NewsData, 0)

	for _, newsID := range newsIDList {
		newsData, err := c.GetNews(newsID)
		if err != nil {
			return nil, err
		}
		newsDataList = append(newsDataList, newsData)
	}

	return newsDataList, nil
}

func (c *CtiNewsSpider) GetNewsIdList() ([]string, error) {
	// 建立新的收集器
	collector := colly.NewCollector()

	// 儲存新聞ID列表
	var newsIDs []string

	// 設定請求頭
	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
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

	log.Printf("找到 %d 篇新聞文章", len(newsIDs))
	return newsIDs, nil
}

// TODO: rename func
func (c *CtiNewsSpider) ArticleScrapingHandle(ctx context.Context, msg []byte) error {
	var event GetNewsEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		c.log.Error().Err(err).Msg("failed to unmarshal message to GetNewsEvent")
		return err
	}

	c.log.Info().Msgf("Processed GetNewsEvent: %+v", event)

	c.log.Info().Msgf("Received message: %s", string(msg))

	// get news id list
	newsIDList, err := c.GetNewsIdList()
	if err != nil {
		c.log.Error().Err(err).Msg("failed to get news id list")
		return err
	}

	c.log.Info().Msgf("Found %d news", len(newsIDList))

	// check id exists

	// get news

	// save news
	return nil
}
