package service

import (
	"context"

	"itmrchow/tw-media-analytics-service/domain/utils"
)

type NewsService interface {

	// 檢查新聞是否存在
	CheckNewsExist(ctx context.Context, checkNews utils.EventNewsCheck) (err error)

	// 保存新聞
	SaveNews(ctx context.Context, saveNews utils.EventNewsSave) error
}
