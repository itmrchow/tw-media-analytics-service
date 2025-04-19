package usecase

import (
	"itmrchow/tw-media-analytics-service/domain/spider/entity"
)

type Spider interface {
	// 爬取新聞
	GetNews(newsID string) (*entity.News, error)
	// 爬取多個新聞
	GetNewsList(newsIDList []string) ([]*entity.News, error)
	// 爬取新聞ID列表
	GetNewsIdList() ([]string, error)
	// 爬取媒體ID
	GetMediaID() uint
}
