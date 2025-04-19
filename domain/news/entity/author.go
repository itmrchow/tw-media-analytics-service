package entity

import (
	"gorm.io/gorm"
)

// Author represents an author entity
type Author struct {
	gorm.Model
	Name    string `gorm:"type:varchar(255);not null"`
	MediaID uint   `gorm:"not null"`

	Media    Media  `gorm:"foreignKey:MediaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	NewsList []News `gorm:"foreignKey:AuthorID"`
}
