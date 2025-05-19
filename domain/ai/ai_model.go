package ai

import "itmrchow/tw-media-analytics-service/domain/ai/dto"

type AiModel interface {
	AnalyzeNews(title string, content string) (*dto.NewsAnalytics, error)
	CloseClient() error
}
