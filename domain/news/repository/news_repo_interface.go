package repository

import "itmrchow/tw-media-analytics-service/domain/news/entity"

type NewsRepository interface {
	BaseRepository[NewsRepository]
	// CreateNews(news *entity.News) error

	// FindNonExistingNewsIDs 根據媒體ID和新聞ID列表，找出在資料庫中不存在的新聞ID
	// Args:
	//   mediaID: 媒體ID
	//   newsIDList: 新聞ID列表
	// Returns:
	//   []string: 不存在的新聞ID列表
	//   error: 錯誤資訊
	FindNonExistingNewsIDs(mediaID uint, newsIDList []string) ([]string, error)
	SaveNews(news *entity.News) error
}
