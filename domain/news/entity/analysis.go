package entity

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Analysis struct {
	gorm.Model
	NewsID  string          `json:"news_id" gorm:"type:char(36);not null;;uniqueIndex:idx_news_media_type"`
	MediaID string          `json:"media_id" gorm:"type:char(36);not null;;uniqueIndex:idx_news_media_type"`
	Type    AnalysisType    `json:"type" gorm:"type:varchar(255);not null;;uniqueIndex:idx_news_media_type"`
	Score   decimal.Decimal `json:"score" gorm:"type:decimal(10,2);not null"`
	Reason  string          `json:"reason" gorm:"type:text;not null"`

	// Relations
	News                News             `gorm:"foreignKey:NewsID,MediaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AnalysisMetricsList []AnalysisMetric `gorm:"foreignKey:AnalysisID"`
}

type AnalysisType string

const (
	AnalysisTypeContent AnalysisType = "content"
	AnalysisTypeTitle   AnalysisType = "title"
)
