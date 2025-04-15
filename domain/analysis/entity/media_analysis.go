package entity

import "github.com/shopspring/decimal"

type MediaAnalysis struct {
	AnalyticsID string
	MediaID     string
	NewsID      string

	TitleScore          decimal.Decimal
	TitleReason         string
	TitleIsNameIncluded bool

	ContentScore          decimal.Decimal
	ContentReason         string
	ContentIsNameIncluded bool
}
