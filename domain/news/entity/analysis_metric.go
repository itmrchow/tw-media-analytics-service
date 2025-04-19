package entity

import (
	"github.com/shopspring/decimal"

	"itmrchow/tw-media-analytics-service/domain/utils"
)

type AnalysisMetric struct {
	utils.TimeModel
	AnalysisID string          `json:"analysis_id" gorm:"primaryKey;type:char(36);not null;index"`
	MetricKey  string          `json:"metric_key" gorm:"primaryKey;type:varchar(255);not null;index"`
	Score      decimal.Decimal `json:"score" gorm:"type:decimal(10,2);not null"`
	Reason     string          `json:"reason" gorm:"type:text;not null"`

	Analysis Analysis `gorm:"foreignKey:AnalysisID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type AnalysisMetricKey string

const (
	MetricKeyTitleAccuracy       AnalysisMetricKey = "accuracy"       // 標題準確性
	MetricKeyTitleClarity        AnalysisMetricKey = "clarity"        // 標題清晰度
	MetricKeyTitleObjectivity    AnalysisMetricKey = "objectivity"    // 標題客觀性
	MetricKeyTitleRelevance      AnalysisMetricKey = "relevance"      // 標題相關性
	MetricKeyTitleAttractiveness AnalysisMetricKey = "attractiveness" // 標題吸引力

	MetricKeyContentAccuracy     AnalysisMetricKey = "accuracy"     // 內容準確性
	MetricKeyContentObjectivity  AnalysisMetricKey = "objectivity"  // 內容客觀性
	MetricKeyContentTimeliness   AnalysisMetricKey = "timeliness"   // 內容即時性
	MetricKeyContentImportance   AnalysisMetricKey = "importance"   // 內容重要性
	MetricKeyContentPresentation AnalysisMetricKey = "presentation" // 內容呈現性
)
