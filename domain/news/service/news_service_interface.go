package service

import "context"

type NewsService interface {
	// NewsService interface method to find non-existing news IDs in the database
	// FindNonExistingNewsIDs(mediaName string, newsIDList []string) ([]string, error)

	// CreateNews(news *News) error

	// 檢查新聞是否存在 , sub handler
	CheckNewsExistHandle(ctx context.Context, msg []byte) error

	// 保存新聞 , sub handler
	SaveNewsHandle(ctx context.Context, msg []byte) error
}
