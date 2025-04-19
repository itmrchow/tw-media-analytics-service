package utils

import "time"

type EventArticleListScraping struct {
	NewsIDList []string
}

type EventNewsCheck struct {
	MediaID    uint
	NewsIDList []string
}

type EventTopicArticleContentScraping struct {
	MediaID uint
	NewsID  string
}

type EventNewsSave struct {
	MediaID     uint
	NewsID      string
	Title       string
	Content     string
	URL         string
	AuthorName  string
	PublishedAt time.Time
	Category    string
}

type EventGetAnalysis struct {
	MediaID uint
	NewsID  string
}

// TODO: 分析保存struct
type EventAnalysisSave struct {
	NewsID   string
	Analysis string
}
