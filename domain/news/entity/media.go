package entity

import (
	"gorm.io/gorm"
)

// Media represents a media source entity
type Media struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null"`

	// Relations
	NewsList   []News   `gorm:"foreignKey:MediaID"`
	AuthorList []Author `gorm:"foreignKey:MediaID"`
}

type MediaID uint

const (
	MediaIDCtiNews  MediaID = iota + 1 // 中天
	MediaIDSetnNews                    // 三立
)
