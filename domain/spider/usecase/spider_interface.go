package usecase

import (
	"itmrchow/tw-media-analytics-service/domain/spider/entity"
)

type Spider interface {
	GetNews(newsID string) (*entity.News, error)
	GetNewsList(newsIDList []string) ([]*entity.News, error)
	GetNewsIdList() ([]string, error)
	GetMediaID() uint
}
