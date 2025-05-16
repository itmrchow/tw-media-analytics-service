package repository

import (
	"os"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"itmrchow/tw-media-analytics-service/domain/news/entity"
	"itmrchow/tw-media-analytics-service/infra"
)

func TestAnalysisRepoSuite(t *testing.T) {
	suite.Run(t, new(AnalysisTestSuite))
}

type AnalysisTestSuite struct {
	suite.Suite
	analysisRepo AnalysisRepository
	db           *gorm.DB
	log          *zerolog.Logger
}

func (s *AnalysisTestSuite) SetupTest() {
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	s.log = &logger
	s.db = infra.InitSqliteDb()

	sqlDB, err := s.db.DB()
	s.Require().NoError(err)

	// init test data
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("sqlite"),
		testfixtures.Directory("testdata"),
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	s.Require().NoError(err)
	err = fixtures.Load()
	s.Require().NoError(err)

	s.analysisRepo = NewAnalysisRepositoryImpl(s.log, s.db)
}

func (s *AnalysisTestSuite) TestSaveAnalysisList() {
	tests := []struct {
		name         string
		analysisList []entity.Analysis
		wantErr      bool
	}{
		{
			name: "成功保存分析列表",
			analysisList: []entity.Analysis{
				{
					NewsID:  "test-news-1",
					MediaID: 1,
					Type:    entity.AnalysisTypeTitle,
					Score:   decimal.NewFromFloat(0.85),
					Reason:  "測試原因",
					AnalysisMetricsList: []entity.AnalysisMetric{
						{
							MetricKey: string(entity.MetricKeyTitleAccuracy),
							Score:     decimal.NewFromFloat(0.9),
							Reason:    "準確度測試",
						},
						{
							MetricKey: string(entity.MetricKeyTitleClarity),
							Score:     decimal.NewFromFloat(0.8),
							Reason:    "清晰度測試",
						},
					},
				},
				{
					NewsID:  "test-news-2",
					MediaID: 1,
					Type:    entity.AnalysisTypeContent,
					Score:   decimal.NewFromFloat(0.75),
					Reason:  "內容測試原因",
					AnalysisMetricsList: []entity.AnalysisMetric{
						{
							MetricKey: string(entity.MetricKeyContentAccuracy),
							Score:     decimal.NewFromFloat(0.7),
							Reason:    "內容準確度測試",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "空列表測試",
			analysisList: []entity.Analysis{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.analysisRepo.SaveAnalysisList(tt.analysisList)
			if tt.wantErr {
				s.Error(err)
				return
			}

			s.NoError(err)

			// Verify data was saved correctly
			if len(tt.analysisList) > 0 {
				for _, expectedAnalysis := range tt.analysisList {
					var savedAnalysis entity.Analysis
					err = s.db.First(&savedAnalysis, "news_id = ? AND media_id = ? AND type = ?",
						expectedAnalysis.NewsID,
						expectedAnalysis.MediaID,
						expectedAnalysis.Type).Error
					s.NoError(err)

					// Verify analysis fields
					s.Equal(expectedAnalysis.NewsID, savedAnalysis.NewsID)
					s.Equal(expectedAnalysis.MediaID, savedAnalysis.MediaID)
					s.Equal(expectedAnalysis.Type, savedAnalysis.Type)
					s.Equal(expectedAnalysis.Score.String(), savedAnalysis.Score.String())
					s.Equal(expectedAnalysis.Reason, savedAnalysis.Reason)

					// Verify metrics
					var metrics []entity.AnalysisMetric
					err = s.db.Where("analysis_id = ?", savedAnalysis.ID).Find(&metrics).Error
					s.NoError(err)
					s.Equal(len(expectedAnalysis.AnalysisMetricsList), len(metrics))

					for i, expectedMetric := range expectedAnalysis.AnalysisMetricsList {
						s.Equal(expectedMetric.MetricKey, metrics[i].MetricKey)
						s.Equal(expectedMetric.Score.String(), metrics[i].Score.String())
						s.Equal(expectedMetric.Reason, metrics[i].Reason)
					}
				}
			}
		})
	}
}
