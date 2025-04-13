package spider

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

var _ Spider = &SetnSpider{}

// 三立
// 從https://www.setn.com/sitemapGoogleNews.xml 抓列表
// 從列表中抓取新聞

type SetnSpider struct {
	newsPageURL     string
	newsListPageURL string
	goquerySelector string
}

func NewSetnSpider() *SetnSpider {
	var spider = &SetnSpider{
		newsPageURL:     "https://www.setn.com/News.aspx?NewsID=%d",
		newsListPageURL: "https://www.setn.com/sitemapGoogleNews.xml",
		goquerySelector: "script[type='application/ld+json']",
	}

	return spider
}

func (s *SetnSpider) GetNews(newsID int) (*NewsData, error) {

	// 建立新的收集器
	c := colly.NewCollector()

	// 設定請求頭
	c.OnRequest(func(r *colly.Request) {
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
	c.OnResponse(func(r *colly.Response) {
		responseSize = len(r.Body)
	})

	// 處理錯誤
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: %v", err)
	})

	// 處理 HTML - 獲取新聞內容
	c.OnHTML("div#ckuse div#Content1", func(e *colly.HTMLElement) {
		// 獲取新聞內容
		newsData.NewsContext = strings.TrimSpace(e.Text)
	})

	// 處理 HTML
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		// 解析 JSON
		err := json.Unmarshal([]byte(e.Text), &newsData)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return
		}
	})

	// 開始抓取
	url := fmt.Sprintf("https://www.setn.com/News.aspx?NewsID=%d", newsID)
	newsID--
	err := c.Visit(url)
	if err != nil {
		log.Printf("Error visiting URL: %v , URL: %s", err, url)
		return nil, err
	}

	// 計算執行時間
	elapsedTime := time.Since(startTime)

	// 輸出結果
	fmt.Println("=== 新聞資訊 ===")
	fmt.Printf("新聞ID: %d\n", newsData.NewsID)
	fmt.Printf("標題: %s\n", newsData.Headline)
	fmt.Printf("內容: %s\n", strings.TrimSpace(newsData.NewsContext))
	fmt.Printf("作者類型: %s\n", newsData.Author.Type)
	fmt.Printf("作者名稱: %s\n", newsData.Author.Name)
	fmt.Printf("發布時間: %s\n", newsData.DatePublished.Format("2006-01-02 15:04:05"))
	fmt.Printf("修改時間: %s\n", newsData.DateModified.Format("2006-01-02 15:04:05"))

	fmt.Println("\n=== 執行資訊 ===")

	newsData.ElapsedTime = elapsedTime
	newsData.ResponseSize = responseSize

	fmt.Printf("花費時間: %v\n", elapsedTime)
	fmt.Printf("回應大小: %d bytes\n", responseSize)

	return &newsData, nil
}

func (s *SetnSpider) GetNewsList(newsIDList []int) ([]*NewsData, error) {
	newsDataList := make([]*NewsData, 0)

	for _, newsID := range newsIDList {
		newsData, err := s.GetNews(newsID)
		if err != nil {
			return nil, err
		}
		newsDataList = append(newsDataList, newsData)
	}

	return newsDataList, nil
}

func (s *SetnSpider) GetNewsIdList() ([]int, error) {

	// 建立新的收集器
	c := colly.NewCollector()

	// 儲存新聞ID列表
	var newsIDs []int

	// 設定請求頭
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	})

	// 處理錯誤
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error fetching sitemap: %v", err)
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
				// 轉換為整數
				id, err := strconv.Atoi(newsID)
				if err == nil {
					newsIDs = append(newsIDs, id)
				}
			}
		}
	})

	// 開始抓取
	err := c.Visit("https://www.setn.com/sitemapGoogleNews.xml")
	if err != nil {
		return nil, fmt.Errorf("error visiting sitemap: %v", err)
	}

	log.Printf("Found %d news articles", len(newsIDs))
	return newsIDs, nil
}
