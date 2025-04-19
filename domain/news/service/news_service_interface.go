package service

import (
	"context"

	"itmrchow/tw-media-analytics-service/domain/utils"
)

type NewsService interface {
	// NewsService interface method to find non-existing news IDs in the database
	// FindNonExistingNewsIDs(mediaName string, newsIDList []string) ([]string, error)

	// CreateNews(news *News) error

	// 檢查新聞是否存在 , sub handler
	CheckNewsExist(ctx context.Context, checkNews utils.EventNewsCheck) (NoExistNewsIdList []string, err error)

	// 保存新聞 , sub handler
	SaveNews(ctx context.Context, saveNews utils.EventNewsSave) error
}
