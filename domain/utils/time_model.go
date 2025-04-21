package utils

import (
	"time"

	"gorm.io/gorm"
)

// TimeModel is a base model that includes created and updated timestamps
type TimeModel struct {
	CreatedAt time.Time      `gorm:"<-:create;autoCreateTime;"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
