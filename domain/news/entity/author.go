package entity

import (
	"time"
)

// Author represents a news article author.
type Author struct {
	ID        string    `json:"id" bson:"_id" gorm:"primaryKey;type:char(36)"`
	Name      string    `json:"name" bson:"name" gorm:"type:varchar(255);not null"`
	Email     string    `json:"email" bson:"email" gorm:"type:varchar(255);not null;unique"`
	MediaID   string    `json:"media_id" bson:"media_id" gorm:"type:char(36);not null;index"`
	CreatedAt time.Time `json:"created_at" bson:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}
