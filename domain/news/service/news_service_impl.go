package service

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/ai"
	"itmrchow/tw-media-analytics-service/domain/news/entity"
	"itmrchow/tw-media-analytics-service/domain/news/repository"
	"itmrchow/tw-media-analytics-service/domain/queue"
	"itmrchow/tw-media-analytics-service/domain/utils"
)

var _ NewsService = &NewsServiceImpl{}

type NewsServiceImpl struct {
	logger *zerolog.Logger

	// queue
	queue queue.Queue

	// repo
	newsRepo     repository.NewsRepository
	authorRepo   repository.AuthorRepository
	analysisRepo repository.AnalysisRepository
	// db
	db *gorm.DB
	// ai model
	aiModel ai.AiModel
}

func NewNewsServiceImpl(
	logger *zerolog.Logger,
	newsRepo repository.NewsRepository,
	authorRepo repository.AuthorRepository,
	analysisRepo repository.AnalysisRepository,
	queue queue.Queue,
	db *gorm.DB,
	aiModel ai.AiModel,
) *NewsServiceImpl {
	return &NewsServiceImpl{
		logger:       logger,
		newsRepo:     newsRepo,
		authorRepo:   authorRepo,
		analysisRepo: analysisRepo,
		queue:        queue,
		db:           db,
		aiModel:      aiModel,
	}
}

// 檢查文章sub handler
func (s *NewsServiceImpl) CheckNewsExist(ctx context.Context, checkNews utils.EventNewsCheck) (err error) {

	s.logger.Info().Msg("check news exist start")

	// check news id exist in db
	nonExistingNewsIDs, err := s.newsRepo.FindNonExistingNewsIDs(checkNews.MediaID, checkNews.NewsIDList)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to find non existing news ids")
		return err
	}

	// print log
	s.logger.Info().
		Str("media_id", strconv.Itoa(int(checkNews.MediaID))).
		Uint("news_id_size", uint(len(nonExistingNewsIDs))).
		Msg("check news exist event")

	if len(nonExistingNewsIDs) == 0 {
		s.logger.Info().Msg("no news id to save")
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
			s.logger.Error().Err(err).Msg("failed to publish news save event")
			return nil // 不影響其他新聞爬取
		}
	}

	s.logger.Info().
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
		s.logger.Error().Err(err).Msg("failed to get or create author")
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
		s.logger.Error().Err(err).Msg("failed to save news")
		return err
	}

	s.logger.Info().
		Str("media_id", strconv.Itoa(int(saveNews.MediaID))).
		Str("news_id", news.NewsID).
		Str("title", news.Title[:min(10, len(news.Title))]).
		Msg("save news")

	return nil
}

// 分析新聞sub handler
func (s *NewsServiceImpl) AnalysisNews(ctx context.Context, analysisNews utils.EventNewsAnalysis) error {

	s.logger.Info().Msgf("news analysis start , analysis num: %d", analysisNews.AnalysisNum)

	// get news is not analysis
	nonAnalysisNews, err := s.newsRepo.FindNonAnalysisNews(analysisNews.AnalysisNum)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to find non analysis news")
		return err
	}

	analysisList := []entity.Analysis{}

	for _, news := range nonAnalysisNews {
		s.logger.Info().Msgf("analysis news to ai model: %s", news.Title)

		// send msg to ai model
		analysis, analysisErr := s.aiModel.AnalyzeNews(news.Title, news.Content)
		if analysisErr != nil {
			s.logger.Error().Err(analysisErr).Msg("failed to analyze news")
			continue
		}

		s.logger.Debug().Interface("analysis", analysis).Msg("ai model analysis news")

		// to entity
		titleAnalysis := entity.Analysis{
			NewsID:              news.NewsID,
			MediaID:             news.MediaID,
			Type:                entity.AnalysisTypeTitle,
			Score:               decimal.NewFromFloat(analysis.TitleAnalytics.Score),
			Reason:              analysis.TitleAnalytics.Reason,
			AnalysisMetricsList: []entity.AnalysisMetric{},
		}
		for _, metric := range analysis.TitleAnalytics.MetricList {
			titleAnalysis.AnalysisMetricsList = append(titleAnalysis.AnalysisMetricsList, entity.AnalysisMetric{
				MetricKey: metric.MetricKey,
				Score:     decimal.NewFromFloat(metric.Score),
				Reason:    metric.Reason,
			})
		}
		analysisList = append(analysisList, titleAnalysis)

		contentAnalysis := entity.Analysis{
			NewsID:              news.NewsID,
			MediaID:             news.MediaID,
			Type:                entity.AnalysisTypeContent,
			Score:               decimal.NewFromFloat(analysis.ContentAnalytics.Score),
			Reason:              analysis.ContentAnalytics.Reason,
			AnalysisMetricsList: []entity.AnalysisMetric{},
		}
		for _, metric := range analysis.ContentAnalytics.MetricList {
			contentAnalysis.AnalysisMetricsList = append(contentAnalysis.AnalysisMetricsList, entity.AnalysisMetric{
				MetricKey: metric.MetricKey,
				Score:     decimal.NewFromFloat(metric.Score),
				Reason:    metric.Reason,
			})
		}
		analysisList = append(analysisList, contentAnalysis)
	}

	// save analysis to db

	err = s.analysisRepo.SaveAnalysisList(analysisList)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to save analysis")
		return err
	}

	return nil
}
