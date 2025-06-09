package usecase

import (
	"context"

	"itmrchow/tw-media-analytics-service/domain/spider/entity"
)

type Spider interface {
	// 爬取新聞
	GetNews(ctx context.Context, newsID string) (*entity.News, error)
	// 爬取多個新聞
	GetNewsList(ctx context.Context, newsIDList []string) ([]*entity.News, error)
	// 爬取新聞ID列表
	GetNewsIdList(ctx context.Context) ([]string, error)
	// 爬取媒體ID
	GetMediaID() uint
}
