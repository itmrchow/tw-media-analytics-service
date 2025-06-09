package delivery

import (
	"context"

	"golang.org/x/sync/errgroup"

	"itmrchow/tw-media-analytics-service/domain/queue"
)

// InitNewsConsumer 初始化新聞相關的 consumer.
func InitNewsConsumer(
	ctx context.Context,
	q queue.Queue,
	handler *NewsEventHandler,
) error {
	// 建立 error group
	var group errgroup.Group

	// NewsCheck consumer
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicNewsCheck, "", handler.CheckNewsExistHandle)
	})

	// NewsSave consumer
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicNewsSave, "", handler.SaveNewsHandle)
	})

	// GetAnalysis consumer
	group.Go(func() error {
		return q.Consume(ctx, queue.TopicGetAnalysis, "", handler.GetAnalysisHandle)
	})

	return group.Wait()
}
