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
	TopicArticleListScraping    QueueTopic = "article_list_scraping"    // 文章列表爬取
	TopicNewsCheck              QueueTopic = "news_check"               // 新聞檢查
	TopicArticleContentScraping QueueTopic = "article_content_scraping" // 文章爬取
	TopicNewsSave               QueueTopic = "news_save"                // 新聞保存

	// analysis news flow
	TopicGetAnalysis  QueueTopic = "analysis_get"  // 取得分析
	TopicAnalysisSave QueueTopic = "analysis_save" // 分析保存
)

func GetTopics() []QueueTopic {
	return []QueueTopic{
		TopicArticleListScraping,
		TopicNewsCheck,
		TopicArticleContentScraping,
		TopicNewsSave,
		TopicGetAnalysis,
		TopicAnalysisSave,
	}
}
