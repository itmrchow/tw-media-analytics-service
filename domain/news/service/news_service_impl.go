package service

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

var _ NewsService = &NewsServiceImpl{}

type NewsServiceImpl struct {
	log *zerolog.Logger

	// queue
	queue queue.Queue

	// repo
	newsRepo   repository.NewsRepository
	authorRepo repository.AuthorRepository
	// db
	db *gorm.DB
}

func NewNewsServiceImpl(
	log *zerolog.Logger,
	newsRepo repository.NewsRepository,
	authorRepo repository.AuthorRepository,
	queue queue.Queue,
	db *gorm.DB,
) *NewsServiceImpl {
	return &NewsServiceImpl{
		log:        log,
		newsRepo:   newsRepo,
		authorRepo: authorRepo,
		queue:      queue,
		db:         db,
	}
}

// 檢查文章sub handler
func (s *NewsServiceImpl) CheckNewsExist(ctx context.Context, checkNews utils.EventNewsCheck) (err error) {

	s.log.Info().Msg("check news exist start")

	// check news id exist in db
	nonExistingNewsIDs, err := s.newsRepo.FindNonExistingNewsIDs(checkNews.MediaID, checkNews.NewsIDList)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to find non existing news ids")
		return err
	}

	// print log
	s.log.Info().
		Str("media_id", strconv.Itoa(int(checkNews.MediaID))).
		Uint("news_id_size", uint(len(nonExistingNewsIDs))).
		Msg("check news exist event")

	if len(nonExistingNewsIDs) == 0 {
		s.log.Info().Msg("no news id to save")
		return nil
	}

	// publish
	for _, newsID := range nonExistingNewsIDs {

		scrapingContentEvent := utils.EventArticleContentScraping{
			MediaID: checkNews.MediaID,
			NewsID:  newsID,
		}

		err = s.queue.Publish(ctx, queue.TopicArticleContentScraping, scrapingContentEvent)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to publish news save event")
			return nil // 不影響其他新聞爬取
		}
	}

	s.log.Info().
		Str("media_id", strconv.Itoa(int(checkNews.MediaID))).
		Uint("news_id_size", uint(len(nonExistingNewsIDs))).
		Msg("send article content scraping event")

	return nil
}

// 保存新聞sub handler
func (s *NewsServiceImpl) SaveNews(ctx context.Context, saveNews utils.EventNewsSave) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// get or create author
	author := &entity.Author{
		MediaID: saveNews.MediaID,
		Name:    saveNews.AuthorName,
	}
	if err := s.authorRepo.FirstOrCreate(ctx, author); err != nil {
		s.log.Error().Err(err).Msg("failed to get or create author")
		return err
	}

	// event dto to news entity
	news := &entity.News{
		MediaID:     saveNews.MediaID,
		NewsID:      saveNews.NewsID,
		Title:       saveNews.Title,
		Content:     saveNews.Content,
		URL:         saveNews.URL,
		AuthorID:    author.ID,
		PublishedAt: saveNews.PublishedAt,
		Category:    saveNews.Category,
	}

	// save news
	if err := s.newsRepo.SaveNews(news); err != nil {
		s.log.Error().Err(err).Msg("failed to save news")
		return err
	}

	s.log.Info().
		Str("media_id", strconv.Itoa(int(saveNews.MediaID))).
		Str("news_id", news.NewsID).
		Str("title", news.Title[:min(10, len(news.Title))]).
		Msg("save news")

	return nil
}

// 分析新聞sub handler
func (s *NewsServiceImpl) AnalysisNews(ctx context.Context, analysisNews utils.EventNewsAnalysis) error {

	s.log.Info().Msgf("news analysis start , analysis num: %d", analysisNews.AnalysisNum)

	// get news is not analysis
	nonAnalysisNews, err := s.newsRepo.FindNonAnalysisNews(analysisNews.AnalysisNum)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to find non analysis news")
		return err
	}

	for _, news := range nonAnalysisNews {
		s.log.Info().Msgf("analysis news: %s", news.Title)

		// send msg to ai model

		// get ai model response

		// save analysis to db
	}

	// send msg to ai model

	// get ai model response

	// save analysis to db

	return nil
}
