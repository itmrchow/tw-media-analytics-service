package entity

import "time"

type News struct {
	NewsID        string
	Headline      string    `json:"headline"`
	Author        Author    `json:"author"`
	DatePublished time.Time `json:"datePublished"`
	DateModified  time.Time `json:"dateModified"`
	NewsContext   string    `json:"newsContext"`

	ResponseSize int           `json:"responseSize"`
	ElapsedTime  time.Duration `json:"elapsedTime"`
}
