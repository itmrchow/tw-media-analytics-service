package entity

import (
	"time"

	"itmrchow/tw-media-analytics-service/domain/utils"
)

// News represents a news article entity
type News struct {
	utils.TimeModel
	NewsID      string `gorm:"primaryKey;type:char(36)"`
	MediaID     uint   `gorm:"primaryKey;"`
	Title       string `gorm:"type:varchar(255);not null"`
	Content     string `gorm:"type:text;not null"`
	URL         string `gorm:"type:varchar(255);not null;unique"`
	AuthorID    uint   `gorm:"not null"`
	Category    string `gorm:"type:varchar(255);not null"`
	PublishedAt time.Time

	// Relations
	Media        Media      `gorm:"foreignKey:MediaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Author       Author     `gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AnalysisList []Analysis `gorm:"foreignKey:NewsID,MediaID"`
}
