package repository

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

type NewsRepositoryImpl struct {
	log *zerolog.Logger
	db  *gorm.DB
}

func NewNewsRepositoryImpl(log *zerolog.Logger, db *gorm.DB) *NewsRepositoryImpl {
	return &NewsRepositoryImpl{log: log, db: db}
}

func (r *NewsRepositoryImpl) FindNonExistingNewsIDs(mediaID uint, newsIDList []string) ([]string, error) {
	// 先檢查輸入
	if len(newsIDList) == 0 {
		return []string{}, nil
	}

	// 使用 map 提升查找效率
	existingMap := make(map[string]struct{}, len(newsIDList))
	var existingNewsIDs []string

	if err := r.db.Model(&entity.News{}).
		Where("media_id = ? AND news_id IN (?)", mediaID, newsIDList).
		Pluck("news_id", &existingNewsIDs).Error; err != nil {
		r.log.Error().Err(err).Msg("查詢已存在的新聞ID失敗")
		return nil, err
	}

	// 建立查找 map
	for _, id := range existingNewsIDs {
		existingMap[id] = struct{}{}
	}

	// 使用預先分配的切片
	nonExistingNewsIDs := make([]string, 0, len(newsIDList))
	for _, newsID := range newsIDList {
		if _, exists := existingMap[newsID]; !exists {
			nonExistingNewsIDs = append(nonExistingNewsIDs, newsID)
		}
	}

	return nonExistingNewsIDs, nil
}
