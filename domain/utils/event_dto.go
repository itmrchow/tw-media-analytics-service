package utils

import "time"

type EventNewsListScraping struct {
}

type EventNewsCheck struct {
	MediaID    uint
	NewsIDList []string
}

type EventArticleContentScraping struct {
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

type EventNewsAnalysis struct {
	AnalysisNum uint
}

// TODO: 分析保存struct
type EventAnalysisSave struct {
	NewsID   string
	Analysis string
}
