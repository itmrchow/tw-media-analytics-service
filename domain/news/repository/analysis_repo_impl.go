package repository

import (
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
)

var _ AnalysisRepository = &AnalysisRepositoryImpl{}

type AnalysisRepositoryImpl struct {
	logger *zerolog.Logger
	db     *gorm.DB
}

// NewAnalysisRepositoryImpl creates a new instance of AnalysisRepository
//
// Args:
//
//	log: logger instance for logging operations
//	db: gorm database instance
//
// Returns:
//
//	*AnalysisRepositoryImpl: new repository instance
func NewAnalysisRepositoryImpl(logger *zerolog.Logger, db *gorm.DB) *AnalysisRepositoryImpl {
	return &AnalysisRepositoryImpl{
		logger: logger,
		db:     db,
	}
}

// SaveAnalysisList saves a list of analysis results in a transaction
//
// Args:
//
//	analysisList: slice of Analysis entities to save
//
// Returns:
//
//	error: error if any occurred during the save operation
func (r *AnalysisRepositoryImpl) SaveAnalysisList(analysisList []entity.Analysis) error {
	r.logger.Debug().Int("count", len(analysisList)).Msg("saving analysis list")

	if len(analysisList) == 0 {
		return nil
	}

	// Use transaction to ensure data consistency
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, analysis := range analysisList {
			// Save analysis with its associations
			if err := tx.Create(&analysis).Error; err != nil {
				return fmt.Errorf("failed to create analysis: %w", err)
			}

			// Save metrics using associations
			if err := tx.Model(&analysis).Association("AnalysisMetricsList").Replace(analysis.AnalysisMetricsList); err != nil {
				return fmt.Errorf("failed to save analysis metrics: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		r.logger.Error().Err(err).Msg("failed to save analysis list")
		return fmt.Errorf("failed to save analysis list: %w", err)
	}

	r.logger.Debug().Msg("successfully saved analysis list")
	return nil
}
