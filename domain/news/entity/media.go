package entity

import (
	"gorm.io/gorm"
)

// Media represents a media source entity
type Media struct {
	gorm.Model
	Name string `gorm:"type:varchar(255);not null"`
	URL  string `gorm:"type:varchar(255);not null"`

	// Relations
	News    []News   `gorm:"foreignKey:MediaID"`
	Authors []Author `gorm:"foreignKey:MediaID"`
}
