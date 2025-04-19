package service

import (
	"context"
	"encoding/json"
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
func (s *NewsServiceImpl) CheckNewsExistHandle(ctx context.Context, msg []byte) (err error) {
	// check msg event type
	var createNewsEvent utils.CheckNewsEvent
	if err := json.Unmarshal(msg, &createNewsEvent); err != nil {
		s.log.Error().Err(err).Msg("failed to unmarshal message to CreateNewsEvent")
		return err
	}

	// check news id exist in db
	nonExistingNewsIDs, err := s.newsRepo.FindNonExistingNewsIDs(createNewsEvent.MediaID, createNewsEvent.NewsIDList)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to find non existing news ids")
		return err
	}

	s.log.Info().
		Str("media_id", strconv.Itoa(int(createNewsEvent.MediaID))).
		Uint("news_id_size", uint(len(nonExistingNewsIDs))).
		Msg("check news exist event")

	if len(nonExistingNewsIDs) > 0 {

		err = s.queue.Publish(ctx, queue.TopicArticleContentScraping, createNewsEvent)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to publish news save event")
			return err
		}

		s.log.Info().
			Str("media_id", strconv.Itoa(int(createNewsEvent.MediaID))).
			Uint("news_id_size", uint(len(nonExistingNewsIDs))).
			Msg("send news save")
	}

	return nil
}

// 保存新聞sub handler
func (s *NewsServiceImpl) SaveNewsHandle(ctx context.Context, msg []byte) error {
	// check msg event type
	var saveNewsEvent utils.SaveNewsEvent
	if err := json.Unmarshal(msg, &saveNewsEvent); err != nil {
		s.log.Error().Err(err).Msg("failed to unmarshal message to SaveNewsEvent")
		return err
	}

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
			MediaID:     saveNewsEvent.MediaID,
			NewsID:      saveNewsEvent.NewsID,
			Title:       saveNewsEvent.Title,
			Content:     saveNewsEvent.Content,
			URL:         saveNewsEvent.URL,
			AuthorID:    author.ID,
			PublishedAt: saveNewsEvent.PublishedAt,
			Category:    saveNewsEvent.Category,
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
