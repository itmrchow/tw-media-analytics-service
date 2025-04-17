package entity

import (
	"time"

	"gorm.io/gorm"
)

// TimeModel is a base model that includes created and updated timestamps
type TimeModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
