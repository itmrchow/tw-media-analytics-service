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
	TopicArticleScraping QueueTopic = "article_scraping"
	TopicNewsSave        QueueTopic = "news_save"
	TopicAnalysisSave    QueueTopic = "analysis_save"
	TopicNewsCheck       QueueTopic = "news_check"
)

func GetTopics() []QueueTopic {
	return []QueueTopic{
		TopicArticleScraping,
		TopicNewsSave,
		TopicAnalysisSave,
		TopicNewsCheck,
	}
}
