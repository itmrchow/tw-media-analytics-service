package spider

import "time"

// NewsData 定義新聞資料結構
type NewsData struct {
	NewsID        int
	Headline      string    `json:"headline"`
	Author        Author    `json:"author"`
	DatePublished time.Time `json:"datePublished"`
	DateModified  time.Time `json:"dateModified"`
	NewsContext   string    `json:"newsContext"`

	ResponseSize int           `json:"responseSize"`
	ElapsedTime  time.Duration `json:"elapsedTime"`
}

// Author 定義作者資訊結構
type Author struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}
