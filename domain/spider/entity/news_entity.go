package entity

import "time"

type News struct {
	NewsID        string
	Headline      string    `json:"headline"`
	Author        Author    `json:"author"`
	DatePublished time.Time `json:"datePublished"`
	DateModified  time.Time `json:"dateModified"`
	NewsContext   string    `json:"newsContext"`
	URL           string    `json:"url"`
	Category      string    `json:"articleSection"`

	ResponseSize int           `json:"responseSize"`
	ElapsedTime  time.Duration `json:"elapsedTime"`
}
