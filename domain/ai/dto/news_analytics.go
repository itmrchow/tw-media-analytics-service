package dto

// NewsAnalytics 定義新聞分析的回應格式
// Metric 定義單一評分指標的結構
type Metric struct {
	MetricKey string  `json:"metricKey"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
}

// Analytics 定義分析結果的基本結構
type Analytics struct {
	Score      float64  `json:"score"`
	Reason     string   `json:"reason"`
	MetricList []Metric `json:"metricList"`
}

// NewsAnalytics 定義新聞分析的完整結果
type NewsAnalytics struct {
	TitleAnalytics   Analytics `json:"titleAnalytics"`
	ContentAnalytics Analytics `json:"contentAnalytics"`
}
