package utils

import "time"

type GetNewsEvent struct {
	NewsIDList []string
}

type CheckNewsEvent struct {
	MediaID    uint
	NewsIDList []string
}

type GetNewsContentEvent struct {
	MediaID    uint
	NewsIDList []string
}

type SaveNewsEvent struct {
	MediaID     uint
	NewsID      string
	Title       string
	Content     string
	URL         string
	AuthorName  string
	PublishedAt time.Time
	Category    string
}
