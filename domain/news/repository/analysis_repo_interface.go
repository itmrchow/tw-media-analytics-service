package repository

import "itmrchow/tw-media-analytics-service/domain/news/entity"

type AnalysisRepository interface {
	// SaveAnalysis(analysis *entity.Analysis) error

	SaveAnalysisList(analysisList []entity.Analysis) error
}
