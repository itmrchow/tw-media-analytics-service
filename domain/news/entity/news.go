package entity

import (
	"time"
)

// News represents a news article entity
type News struct {
	TimeModel
	NewsID      string `gorm:"primaryKey;type:char(36)"`
	MediaID     uint   `gorm:"primaryKey;"`
	Title       string `gorm:"type:varchar(255);not null"`
	Content     string `gorm:"type:text;not null"`
	URL         string `gorm:"type:varchar(255);not null;unique"`
	AuthorID    uint   `gorm:"not null"`
	PublishedAt time.Time

	// Relations
	Media    Media      `gorm:"foreignKey:MediaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Author   Author     `gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Analyses []Analysis `gorm:"foreignKey:NewsID,MediaID"`
}
