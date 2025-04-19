package queue

import "context"

type Queue interface {
	InitTopic() error
	Publish(ctx context.Context, topic QueueTopic, message any) error
	Consume(ctx context.Context, topic QueueTopic, subID string, handler func(ctx context.Context, msg []byte) error) error
	CloseClient() error
}

type QueueTopic string

const (
	// scraping news flow
	TopicArticleListScraping QueueTopic = "article_list_scraping" // 文章列表
	TopicNewsCheck           QueueTopic = "news_check"            // 新聞檢查
	TopicArticleScraping     QueueTopic = "article_scraping"      // 文章爬取
	TopicNewsSave            QueueTopic = "news_save"             // 新聞保存

	// analysis news flow
	TopicAnalysisSave QueueTopic = "analysis_save" // 分析保存
)

func GetTopics() []QueueTopic {
	return []QueueTopic{
		TopicArticleScraping,
		TopicNewsSave,
		TopicAnalysisSave,
		TopicNewsCheck,
	}
}
