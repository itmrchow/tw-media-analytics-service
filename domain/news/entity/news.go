package entity

import (
	"time"
)

// News represents a news article in the system.
type News struct {
	ID          string    `json:"id" bson:"_id" gorm:"primaryKey;type:char(36)"`
	Title       string    `json:"title" bson:"title" gorm:"type:varchar(255);not null"`
	Content     string    `json:"content" bson:"content" gorm:"type:text;not null"`
	URL         string    `json:"url" bson:"url" gorm:"type:varchar(255);not null;unique"`
	MediaID     string    `json:"media_id" bson:"media_id" gorm:"type:char(36);not null;index"`
	AuthorID    string    `json:"author_id" bson:"author_id" gorm:"type:char(36);not null;index"`
	PublishDate time.Time `json:"publish_date" bson:"publish_date" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}
