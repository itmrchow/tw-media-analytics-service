package delivery

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

// InitSpiderConsumer 初始化爬蟲相關的 consumer.
func InitSpiderConsumer(
	ctx context.Context,
	q queue.Queue,
	handler *BaseEventHandler,
) error {
	var group errgroup.Group

	// ArticleListScraping consumer
	for mediaID, h := range handler.SpiderMap {
		mediaID := mediaID // 創建新的變數避免閉包問題
		spiderHandler := h

		group.Go(func() error {
			subID := fmt.Sprintf(
				"%s_%s_%v_sub",
				string(queue.TopicArticleListScraping),
				viper.GetString("ENV"),
				mediaID,
			)
			return q.Consume(ctx, queue.TopicArticleListScraping, subID, spiderHandler.ArticleListScrapingHandle)
		})
	}

	// ArticleContentScraping consumer
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicArticleContentScraping, "", handler.ArticleContentScrapingHandle)
	})

	return group.Wait()
}
