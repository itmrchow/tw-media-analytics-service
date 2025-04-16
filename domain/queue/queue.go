package queue

import "context"

type Queue interface {
	CreateTopic() error
	Publish(ctx context.Context, topic QueueTopic, message any) error
	Consume(ctx context.Context, topic QueueTopic, handler func(ctx context.Context, msg []byte) error) error
	CloseClient() error
}

type QueueTopic string

const (
	TopicArticleScraping QueueTopic = "article_scraping"
	TopicNewsSave        QueueTopic = "news_save"
	TopicAnalysisSave    QueueTopic = "analysis_save"
)

func GetTopics() []QueueTopic {
	return []QueueTopic{
		TopicArticleScraping,
		TopicNewsSave,
		TopicAnalysisSave,
	}
}

type MessageHandler interface {
	Handle(message []byte) error
}
