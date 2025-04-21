package ai_model

import "itmrchow/tw-media-analytics-service/domain/ai_model/dto"

type AiModel interface {
	AnalyzeNews(title string, content string) (*dto.NewsAnalytics, error)
	CloseClient() error
}
