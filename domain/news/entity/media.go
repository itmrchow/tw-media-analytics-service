package entity

import (
	"time"
)

// Media represents a media organization that publishes news.
type Media struct {
	ID          string    `json:"id" bson:"_id" gorm:"primaryKey;type:char(36)"`
	Name        string    `json:"name" bson:"name" gorm:"type:varchar(255);not null;unique"`
	Description string    `json:"description" bson:"description" gorm:"type:text"`
	URL         string    `json:"url" bson:"url" gorm:"type:varchar(255);not null;unique"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}
