package service

import (
	"context"
	"strconv"

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
func (s *NewsServiceImpl) CheckNewsExist(ctx context.Context, checkNews utils.EventNewsCheck) (NoExistNewsIdList []string, err error) {

	s.log.Info().Msg("check news exist start")

	// check news id exist in db
	nonExistingNewsIDs, err := s.newsRepo.FindNonExistingNewsIDs(checkNews.MediaID, checkNews.NewsIDList)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to find non existing news ids")
		return nil, err
	}

	s.log.Info().
		Str("media_id", strconv.Itoa(int(checkNews.MediaID))).
		Uint("news_id_size", uint(len(nonExistingNewsIDs))).
		Msg("check news exist event")

	return nonExistingNewsIDs, err

	// if len(nonExistingNewsIDs) > 0 {

	// 	for _, newsID := range nonExistingNewsIDs {

	// 		s.log.Info().Str("news_id", newsID).Msg("")

	// 		scrapingContentEvent := utils.EventTopicArticleContentScraping{
	// 			MediaID: checkNews.MediaID,
	// 			NewsID:  newsID,
	// 		}

	// 		err = s.queue.Publish(ctx, queue.TopicArticleContentScraping, scrapingContentEvent)
	// 		if err != nil {
	// 			s.log.Error().Err(err).Msg("failed to publish news save event")
	// 			return nil // 不影響其他新聞爬取
	// 		}
	// 	}

	// 	s.log.Info().
	// 		Str("media_id", strconv.Itoa(int(checkNews.MediaID))).
	// 		Uint("news_id_size", uint(len(nonExistingNewsIDs))).
	// 		Msg("send news save")
	// }

	// return nil
}

// 保存新聞sub handler
func (s *NewsServiceImpl) SaveNews(ctx context.Context, saveNews utils.EventNewsSave) error {

	// create tx
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		// get tx repo
		authorRepo := s.authorRepo.WithTransaction(tx)
		newsRepo := s.newsRepo.WithTransaction(tx)

		// get or create author
		author := &entity.Author{}
		if err := authorRepo.FirstOrCreate(author); err != nil {
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
		if err := newsRepo.SaveNews(news); err != nil {
			s.log.Error().Err(err).Msg("failed to save news")
			return err
		}

		return nil
	}); err != nil {
		s.log.Error().Err(err).Msg("failed in tx")
		return err
	}

	return nil
}
